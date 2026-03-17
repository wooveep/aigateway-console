#!/usr/bin/env python3

import argparse
import json
import ssl
import subprocess
import sys
import time
import urllib.parse
import urllib.request
from copy import deepcopy
from pathlib import Path


SCRIPT_DIR = Path(__file__).resolve().parent
COMPOSE_DIR = SCRIPT_DIR.parent
EXPORT_DIR = COMPOSE_DIR / ".migration"
PROJECT_NAME = "aigateway"

SOURCE_NAMESPACE = "higress-system"
TARGET_NAMESPACE = "aigateway-system"
OLD_INGRESS_CLASS = "higress"
NEW_INGRESS_CLASS = "aigateway"
OLD_GATEWAY = "higress-gateway"
NEW_GATEWAY = "aigateway-gateway"
OLD_CONSOLE = "higress-console"
NEW_CONSOLE = "aigateway-console"
OLD_SELECTOR_KEY = "higress"
NEW_SELECTOR_KEY = "aigateway"
OLD_SELECTOR_VALUE = f"{SOURCE_NAMESPACE}-{OLD_GATEWAY}"
NEW_SELECTOR_VALUE = f"{TARGET_NAMESPACE}-{NEW_GATEWAY}"
OLD_FILTER = "higress-gateway-global-custom-response"
NEW_FILTER = "aigateway-gateway-global-custom-response"

RESOURCE_SPECS = [
    {
        "name": "gatewayclasses",
        "list_path": "/apis/gateway.networking.k8s.io/v1beta1/gatewayclasses",
        "create_path": "/apis/gateway.networking.k8s.io/v1beta1/gatewayclasses",
        "namespaced": False,
    },
    {
        "name": "configmaps",
        "list_path": "/api/v1/namespaces/{namespace}/configmaps",
        "create_path": "/api/v1/namespaces/{namespace}/configmaps",
        "namespaced": True,
    },
    {
        "name": "secrets",
        "list_path": "/api/v1/namespaces/{namespace}/secrets",
        "create_path": "/api/v1/namespaces/{namespace}/secrets",
        "namespaced": True,
    },
    {
        "name": "services",
        "list_path": "/api/v1/namespaces/{namespace}/services",
        "create_path": "/api/v1/namespaces/{namespace}/services",
        "namespaced": True,
    },
    {
        "name": "mcpbridges",
        "list_path": "/apis/networking.higress.io/v1/namespaces/{namespace}/mcpbridges",
        "create_path": "/apis/networking.higress.io/v1/namespaces/{namespace}/mcpbridges",
        "namespaced": True,
    },
    {
        "name": "envoyfilters",
        "list_path": "/apis/networking.istio.io/v1alpha3/namespaces/{namespace}/envoyfilters",
        "create_path": "/apis/networking.istio.io/v1alpha3/namespaces/{namespace}/envoyfilters",
        "namespaced": True,
    },
    {
        "name": "gateways",
        "list_path": "/apis/gateway.networking.k8s.io/v1beta1/namespaces/{namespace}/gateways",
        "create_path": "/apis/gateway.networking.k8s.io/v1beta1/namespaces/{namespace}/gateways",
        "namespaced": True,
    },
    {
        "name": "wasmplugins",
        "list_path": "/apis/extensions.higress.io/v1alpha1/namespaces/{namespace}/wasmplugins",
        "create_path": "/apis/extensions.higress.io/v1alpha1/namespaces/{namespace}/wasmplugins",
        "namespaced": True,
    },
    {
        "name": "ingresses",
        "list_path": "/apis/networking.k8s.io/v1/namespaces/{namespace}/ingresses",
        "create_path": "/apis/networking.k8s.io/v1/namespaces/{namespace}/ingresses",
        "namespaced": True,
    },
]

NAME_REPLACEMENTS = {
    OLD_GATEWAY: NEW_GATEWAY,
    OLD_CONSOLE: NEW_CONSOLE,
    OLD_FILTER: NEW_FILTER,
}

VALUE_REPLACEMENTS = [
    (f"{OLD_CONSOLE}.{SOURCE_NAMESPACE}.svc.cluster.local", f"{NEW_CONSOLE}.{TARGET_NAMESPACE}.svc.cluster.local"),
    (f"{OLD_GATEWAY}.{SOURCE_NAMESPACE}.svc.cluster.local", f"{NEW_GATEWAY}.{TARGET_NAMESPACE}.svc.cluster.local"),
    (OLD_SELECTOR_VALUE, NEW_SELECTOR_VALUE),
    (OLD_FILTER, NEW_FILTER),
    (f"{OLD_CONSOLE}.dns", f"{NEW_CONSOLE}.dns"),
    (OLD_GATEWAY, NEW_GATEWAY),
    (OLD_CONSOLE, NEW_CONSOLE),
    (SOURCE_NAMESPACE, TARGET_NAMESPACE),
]

HTTP_TIMEOUT = 20
SSL_CONTEXT = ssl._create_unverified_context()


def log(message: str) -> None:
    print(f"[migrate] {message}", flush=True)


def run(cmd, cwd=COMPOSE_DIR, capture=False):
    result = subprocess.run(
        cmd,
        cwd=cwd,
        check=True,
        text=True,
        capture_output=capture,
    )
    return result.stdout.strip() if capture else ""


def compose(*args, capture=False):
    return run(["docker", "compose", "-p", PROJECT_NAME, *args], capture=capture)


def docker_inspect(container_id: str, fmt: str) -> str:
    return run(["docker", "inspect", "--format", fmt, container_id], capture=True)


def http_request(url, method="GET", params=None, json_body=None, headers=None, verify_tls=True):
    headers = headers or {}
    data = None
    if method in {"GET", "DELETE"} and params:
        query = urllib.parse.urlencode(params, doseq=True)
        sep = "&" if "?" in url else "?"
        url = f"{url}{sep}{query}"
    elif params:
        data = urllib.parse.urlencode(params, doseq=True).encode()
        headers.setdefault("Content-Type", "application/x-www-form-urlencoded")

    if json_body is not None:
        data = json.dumps(json_body).encode()
        headers.setdefault("Content-Type", "application/json")

    request = urllib.request.Request(url, data=data, headers=headers, method=method)
    context = SSL_CONTEXT if url.startswith("https://") and not verify_tls else None
    with urllib.request.urlopen(request, timeout=HTTP_TIMEOUT, context=context) as response:
        body = response.read().decode()
        content_type = response.headers.get("Content-Type", "")
        if "application/json" in content_type:
            return json.loads(body)
        return body


def compose_container_id(service: str) -> str:
    container_id = run(
        [
            "sh",
            "-lc",
            "docker ps "
            f"--filter label=com.docker.compose.project={PROJECT_NAME} "
            f"--filter label=com.docker.compose.service={service} "
            "-q | head -n 1",
        ],
        capture=True,
    )
    if not container_id:
        raise RuntimeError(f"service '{service}' is not available")
    return container_id


def apiserver_base_url() -> str:
    container_id = compose_container_id("apiserver")
    ip_address = docker_inspect(container_id, "{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}")
    if not ip_address:
        raise RuntimeError("failed to resolve apiserver container IP")
    return f"https://{ip_address}:8443"


def wait_for_service(service: str, timeout_seconds: int = 180) -> None:
    deadline = time.time() + timeout_seconds
    while time.time() < deadline:
        try:
            container_id = compose_container_id(service)
            status = docker_inspect(
                container_id,
                "{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}",
            )
            if status in {"healthy", "running", "exited"}:
                return
        except Exception:
            pass
        time.sleep(2)
    raise RuntimeError(f"service '{service}' did not become ready in time")


def wait_for_apiserver(timeout_seconds: int = 180) -> None:
    deadline = time.time() + timeout_seconds
    while time.time() < deadline:
        try:
            body = http_request(f"{apiserver_base_url()}/readyz", verify_tls=False)
            if str(body).strip() == "ok":
                return
        except Exception:
            pass
        time.sleep(2)
    raise RuntimeError("apiserver did not become ready in time")


def nacos_namespaces():
    payload = http_request("http://127.0.0.1:8888/v1/console/namespaces")
    return payload.get("data", payload)


def ensure_target_namespace(target_namespace: str) -> None:
    namespaces = {item["namespace"]: item for item in nacos_namespaces()}
    if target_namespace in namespaces:
        return
    log(f"creating Nacos namespace '{target_namespace}'")
    http_request(
        "http://127.0.0.1:8888/v1/console/namespaces",
        method="POST",
        params={
            "customNamespaceId": target_namespace,
            "namespaceName": target_namespace,
            "namespaceDesc": "migrated from higress-system",
        },
    )


def list_nacos_configs(namespace_id: str):
    payload = http_request(
        "http://127.0.0.1:8848/nacos/v1/cs/configs",
        params={
            "search": "blur",
            "dataId": "",
            "group": "",
            "pageNo": 1,
            "pageSize": 500,
            "tenant": namespace_id,
        },
    )
    return payload


def assert_target_empty(target_namespace: str) -> None:
    payload = list_nacos_configs(target_namespace)
    if int(payload.get("totalCount", 0)) != 0:
        raise RuntimeError(
            f"target namespace '{target_namespace}' is not empty; "
            f"found {payload.get('totalCount')} config entries"
        )


def export_resources(source_namespace: str):
    api_base = apiserver_base_url()
    exported = []
    for spec in RESOURCE_SPECS:
        list_path = spec["list_path"].format(namespace=source_namespace)
        payload = http_request(f"{api_base}{list_path}", verify_tls=False)
        items = payload.get("items", [])
        log(f"exported {len(items)} {spec['name']}")
        exported.append({"spec": spec, "items": items})
    return exported


def transform_label_map(mapping):
    if not isinstance(mapping, dict):
        return mapping
    result = {}
    for key, value in mapping.items():
        new_key = NEW_SELECTOR_KEY if key == OLD_SELECTOR_KEY else key
        if key == "app":
            value = transform_string(value, ("metadata", "labels", "app"))
        elif key == OLD_SELECTOR_KEY:
            value = NEW_SELECTOR_VALUE if value == OLD_SELECTOR_VALUE else transform_string(
                value, ("metadata", "labels", NEW_SELECTOR_KEY)
            )
        result[new_key] = value
    return result


def transform_string(value: str, path) -> str:
    if not isinstance(value, str):
        return value
    if path and path[-1] == "controllerName" and value == "higress.io/gateway-controller":
        return value
    if len(path) >= 3 and path[-3:] == ("metadata", "labels", "higress.io/resource-definer"):
        return value
    for old, new in VALUE_REPLACEMENTS:
        value = value.replace(old, new)
    if path and path[-1] == "ingressClassName" and value == OLD_INGRESS_CLASS:
        return NEW_INGRESS_CLASS
    return value


def transform_any(value, path=()):
    if isinstance(value, dict):
        transformed = {}
        for key, child in value.items():
            next_key = NEW_SELECTOR_KEY if key == OLD_SELECTOR_KEY and path and path[-1] in {
                "labels",
                "selector",
                "matchLabels",
            } else key
            transformed[next_key] = transform_any(child, path + (next_key,))
        return transformed
    if isinstance(value, list):
        return [transform_any(item, path + ("[]",)) for item in value]
    if isinstance(value, str):
        return transform_string(value, path)
    return value


def sanitize_metadata(metadata: dict, namespaced: bool) -> dict:
    metadata = deepcopy(metadata)
    for field in ("resourceVersion", "creationTimestamp", "managedFields", "uid", "generation", "selfLink"):
        metadata.pop(field, None)
    if namespaced:
        metadata["namespace"] = TARGET_NAMESPACE
    name = metadata.get("name")
    if name in NAME_REPLACEMENTS:
        metadata["name"] = NAME_REPLACEMENTS[name]
    if "labels" in metadata:
        metadata["labels"] = transform_label_map(metadata["labels"])
    if "annotations" in metadata and isinstance(metadata["annotations"], dict):
        metadata["annotations"] = {
            key: transform_string(value, ("metadata", "annotations", key))
            for key, value in metadata["annotations"].items()
        }
    return metadata


def sanitize_spec_fields(kind: str, spec: dict) -> dict:
    spec = deepcopy(spec)
    if kind == "Service":
        spec.pop("clusterIP", None)
        spec.pop("clusterIPs", None)
        spec.pop("internalTrafficPolicy", None)
        spec.pop("ipFamilies", None)
        spec.pop("ipFamilyPolicy", None)
        spec.pop("sessionAffinityConfig", None)
        selector = spec.get("selector")
        if isinstance(selector, dict):
            spec["selector"] = transform_label_map(selector)
    elif kind == "Ingress":
        spec["ingressClassName"] = NEW_INGRESS_CLASS
    elif kind == "Gateway":
        if spec.get("gatewayClassName") == OLD_GATEWAY:
            spec["gatewayClassName"] = NEW_GATEWAY
    return spec


def transform_item(item: dict, namespaced: bool) -> dict:
    item = deepcopy(item)
    kind = item.get("kind", "")
    item.pop("status", None)
    item["metadata"] = sanitize_metadata(item.get("metadata", {}), namespaced)
    if "spec" in item and isinstance(item["spec"], dict):
        item["spec"] = sanitize_spec_fields(kind, item["spec"])
    item = transform_any(item)

    if kind == "GatewayClass":
        item["metadata"]["name"] = NEW_GATEWAY if item["metadata"]["name"] == OLD_GATEWAY else item["metadata"]["name"]
    elif kind == "Service":
        item["metadata"]["name"] = NEW_GATEWAY if item["metadata"]["name"] == OLD_GATEWAY else item["metadata"]["name"]
    elif kind == "Secret":
        item["type"] = item.get("type", "Opaque")
    return item


def transform_resources(exported):
    transformed = []
    for group in exported:
        spec = group["spec"]
        items = [transform_item(item, spec["namespaced"]) for item in group["items"]]
        transformed.append({"spec": spec, "items": items})
    return transformed


def write_export(bundle, export_file: Path) -> None:
    export_file.parent.mkdir(parents=True, exist_ok=True)
    export_file.write_text(json.dumps(bundle, ensure_ascii=True, indent=2))


def import_resources(bundle):
    api_base = apiserver_base_url()
    for group in bundle:
        spec = group["spec"]
        create_path = spec["create_path"].format(namespace=TARGET_NAMESPACE)
        url = f"{api_base}{create_path}"
        for item in group["items"]:
            name = item.get("metadata", {}).get("name", "<unknown>")
            http_request(url, method="POST", json_body=item, verify_tls=False)
            log(f"imported {spec['name']}/{name}")


def delete_source_namespace(source_namespace: str) -> None:
    payload = list_nacos_configs(source_namespace)
    for item in payload.get("pageItems", []):
        http_request(
            "http://127.0.0.1:8848/nacos/v1/cs/configs",
            method="DELETE",
            params={
                "dataId": item["dataId"],
                "group": item["group"],
                "tenant": source_namespace,
            },
        )
        log(f"deleted source config {item['group']}/{item['dataId']}")
    http_request(
        "http://127.0.0.1:8888/v1/console/namespaces",
        method="DELETE",
        params={"namespaceId": source_namespace},
    )
    log(f"deleted source namespace '{source_namespace}'")


def verify(target_namespace: str) -> None:
    payload = list_nacos_configs(target_namespace)
    total_count = int(payload.get("totalCount", 0))
    log(f"target namespace '{target_namespace}' now has {total_count} config entries")

    login_response = subprocess.run(
        [
            "curl",
            "-sS",
            "-c",
            "/tmp/aigateway-console-cookie.txt",
            "-H",
            "Content-Type: application/json",
            "-d",
            '{"username":"admin","password":"admin"}',
            "http://127.0.0.1:8080/session/login",
        ],
        check=True,
        text=True,
        capture_output=True,
    )
    if not login_response.stdout.strip():
        raise RuntimeError("console login returned an empty response")

    services = subprocess.run(
        [
            "curl",
            "-sS",
            "-b",
            "/tmp/aigateway-console-cookie.txt",
            "http://127.0.0.1:8080/v1/services",
        ],
        check=True,
        text=True,
        capture_output=True,
    )
    payload = json.loads(services.stdout)
    items = payload.get("data", [])
    if any(item.get("name") == "aigateway-console.dns" for item in items):
        log("console service list contains aigateway-console.dns")
    else:
        raise RuntimeError("console service list does not contain aigateway-console.dns")


def migrate(args) -> None:
    source_payload = list_nacos_configs(args.source_namespace)
    source_count = int(source_payload.get("totalCount", 0))
    if source_count == 0:
        raise RuntimeError(f"source namespace '{args.source_namespace}' has no configs to migrate")

    ensure_target_namespace(args.target_namespace)
    assert_target_empty(args.target_namespace)

    exported = export_resources(args.source_namespace)
    transformed = transform_resources(exported)
    write_export(transformed, args.export_file)
    log(f"wrote transformed export to {args.export_file}")

    log("stopping controller, pilot, gateway, and console for the namespace switch")
    compose("stop", "console", "gateway", "pilot", "controller")

    log("recreating apiserver with the new Nacos namespace")
    compose("up", "-d", "--force-recreate", "apiserver")
    wait_for_apiserver()

    log("importing transformed resources into the new namespace")
    import_resources(transformed)

    log("running prepare to regenerate local volumes from the new config namespace")
    compose("up", "--force-recreate", "--no-deps", "prepare")

    for service in ("controller", "pilot", "gateway", "console"):
        log(f"starting {service}")
        compose("up", "-d", "--no-deps", "--force-recreate", service)
        wait_for_service(service)

    log("refreshing observability containers")
    compose("up", "-d", "--no-deps", "--force-recreate", "prometheus", "promtail", "grafana")

    verify(args.target_namespace)

    if args.keep_source:
        log("keeping source namespace as requested")
    else:
        delete_source_namespace(args.source_namespace)


def parse_args():
    parser = argparse.ArgumentParser(description="Migrate standalone Higress Nacos data to the aigateway namespace.")
    parser.add_argument("--source-namespace", default=SOURCE_NAMESPACE)
    parser.add_argument("--target-namespace", default=TARGET_NAMESPACE)
    parser.add_argument(
        "--export-file",
        default=str(EXPORT_DIR / "aigateway-nacos-export.json"),
        type=Path,
        help="where to write the transformed export bundle",
    )
    parser.add_argument(
        "--keep-source",
        action="store_true",
        help="keep the old higress-system namespace as a rollback backup",
    )
    return parser.parse_args()


def main():
    args = parse_args()
    try:
        migrate(args)
    except Exception as exc:
        log(f"migration failed: {exc}")
        sys.exit(1)


if __name__ == "__main__":
    main()
