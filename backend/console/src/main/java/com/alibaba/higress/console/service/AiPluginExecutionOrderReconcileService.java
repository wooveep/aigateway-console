package com.alibaba.higress.console.service;

import java.util.Arrays;
import java.util.List;
import java.util.Objects;
import java.util.stream.Collectors;

import javax.annotation.PostConstruct;
import javax.annotation.Resource;

import org.apache.commons.collections4.CollectionUtils;
import org.apache.commons.collections4.MapUtils;
import org.apache.commons.lang3.ObjectUtils;
import org.springframework.stereotype.Service;

import com.alibaba.higress.sdk.constant.KubernetesConstants;
import com.alibaba.higress.sdk.constant.plugin.BuiltInPluginName;
import com.alibaba.higress.sdk.model.WasmPlugin;
import com.alibaba.higress.sdk.service.WasmPluginService;
import com.alibaba.higress.sdk.service.kubernetes.KubernetesClientService;
import com.alibaba.higress.sdk.service.kubernetes.KubernetesModelConverter;
import com.alibaba.higress.sdk.service.kubernetes.crd.wasm.PluginPhase;
import com.alibaba.higress.sdk.service.kubernetes.crd.wasm.V1alpha1WasmPlugin;

import io.kubernetes.client.openapi.ApiException;
import lombok.extern.slf4j.Slf4j;

@Slf4j
@Service
public class AiPluginExecutionOrderReconcileService {

    private static final List<String> TARGET_PLUGIN_NAMES = Arrays.asList(
        BuiltInPluginName.AI_STATISTICS,
        BuiltInPluginName.AI_DATA_MASKING
    );

    private WasmPluginService wasmPluginService;
    private KubernetesClientService kubernetesClientService;
    private KubernetesModelConverter kubernetesModelConverter;

    @Resource
    public void setWasmPluginService(WasmPluginService wasmPluginService) {
        this.wasmPluginService = wasmPluginService;
    }

    @Resource
    public void setKubernetesClientService(KubernetesClientService kubernetesClientService) {
        this.kubernetesClientService = kubernetesClientService;
    }

    @Resource
    public void setKubernetesModelConverter(KubernetesModelConverter kubernetesModelConverter) {
        this.kubernetesModelConverter = kubernetesModelConverter;
    }

    @PostConstruct
    public void init() {
        syncNow();
    }

    public synchronized void syncNow() {
        for (String pluginName : TARGET_PLUGIN_NAMES) {
            reconcilePluginOrder(pluginName);
        }
    }

    private void reconcilePluginOrder(String pluginName) {
        WasmPlugin desiredPlugin = wasmPluginService.query(pluginName, null);
        if (desiredPlugin == null || !Boolean.TRUE.equals(desiredPlugin.getBuiltIn())) {
            return;
        }

        try {
            reconcileTargetVersion(pluginName, desiredPlugin);
        } catch (Exception ex) {
            log.warn("Failed to reconcile built-in plugin {} execution order.", pluginName, ex);
        }
    }

    private void reconcileTargetVersion(String pluginName, WasmPlugin desiredPlugin) throws ApiException {
        List<V1alpha1WasmPlugin> existedCrs = listBuiltInPluginCrs(pluginName);
        List<V1alpha1WasmPlugin> currentVersionCrs = filterByVersion(existedCrs, desiredPlugin.getPluginVersion());
        List<V1alpha1WasmPlugin> legacyCrs = filterLegacyCrs(existedCrs, desiredPlugin.getPluginVersion());

        if (CollectionUtils.isEmpty(currentVersionCrs) || needsOrderUpdate(currentVersionCrs, desiredPlugin)) {
            WasmPlugin updateRequest = buildUpdateRequest(currentVersionCrs, desiredPlugin);
            wasmPluginService.updateBuiltIn(updateRequest);
            log.info(
                "Reconciled built-in plugin {} execution order to phase={}, priority={}.",
                pluginName,
                desiredPlugin.getPhase(),
                desiredPlugin.getPriority()
            );
            existedCrs = listBuiltInPluginCrs(pluginName);
            currentVersionCrs = filterByVersion(existedCrs, desiredPlugin.getPluginVersion());
            legacyCrs = filterLegacyCrs(existedCrs, desiredPlugin.getPluginVersion());
        }

        if (CollectionUtils.isEmpty(currentVersionCrs)) {
            log.warn("No target built-in WasmPlugin CR is found for {} with version {}.", pluginName,
                desiredPlugin.getPluginVersion());
            return;
        }

        if (CollectionUtils.isEmpty(legacyCrs)) {
            return;
        }

        V1alpha1WasmPlugin targetCr = currentVersionCrs.stream().filter(Objects::nonNull).findFirst().orElse(null);
        if (targetCr == null) {
            return;
        }

        migrateLegacyConfig(pluginName, targetCr, legacyCrs);
        deleteLegacyCrs(pluginName, legacyCrs);
    }

    private List<V1alpha1WasmPlugin> listBuiltInPluginCrs(String pluginName) throws ApiException {
        return kubernetesClientService.listWasmPlugin(pluginName, null, true);
    }

    private List<V1alpha1WasmPlugin> filterByVersion(List<V1alpha1WasmPlugin> existedCrs, String pluginVersion) {
        return existedCrs.stream()
            .filter(Objects::nonNull)
            .filter(cr -> Objects.equals(pluginVersion, getPluginVersion(cr)))
            .collect(Collectors.toList());
    }

    private List<V1alpha1WasmPlugin> filterLegacyCrs(List<V1alpha1WasmPlugin> existedCrs, String pluginVersion) {
        return existedCrs.stream()
            .filter(Objects::nonNull)
            .filter(cr -> !Objects.equals(pluginVersion, getPluginVersion(cr)))
            .collect(Collectors.toList());
    }

    private void migrateLegacyConfig(String pluginName, V1alpha1WasmPlugin targetCr, List<V1alpha1WasmPlugin> legacyCrs) {
        boolean changed = false;
        for (V1alpha1WasmPlugin legacyCr : legacyCrs) {
            if (legacyCr == null || legacyCr.getSpec() == null) {
                continue;
            }
            kubernetesModelConverter.mergeWasmPluginSpec(legacyCr, targetCr);
            changed = true;
        }

        if (!changed) {
            return;
        }

        try {
            kubernetesClientService.replaceWasmPlugin(targetCr);
            log.info(
                "Migrated {} legacy built-in WasmPlugin config(s) into {}.",
                legacyCrs.size(),
                targetCr.getMetadata() != null ? targetCr.getMetadata().getName() : pluginName
            );
        } catch (ApiException ex) {
            throw new RuntimeException("Failed to migrate legacy config for built-in plugin " + pluginName, ex);
        }
    }

    private void deleteLegacyCrs(String pluginName, List<V1alpha1WasmPlugin> legacyCrs) {
        for (V1alpha1WasmPlugin legacyCr : legacyCrs) {
            String legacyCrName = legacyCr.getMetadata() != null ? legacyCr.getMetadata().getName() : null;
            if (legacyCrName == null) {
                continue;
            }
            try {
                kubernetesClientService.deleteWasmPlugin(legacyCrName);
                log.info("Deleted legacy built-in WasmPlugin CR {} for {}.", legacyCrName, pluginName);
            } catch (ApiException ex) {
                log.warn("Failed to delete legacy built-in WasmPlugin CR {} for {}.", legacyCrName, pluginName, ex);
            }
        }
    }

    private boolean needsOrderUpdate(List<V1alpha1WasmPlugin> existedCrs, WasmPlugin desiredPlugin) {
        if (CollectionUtils.isEmpty(existedCrs)) {
            return true;
        }

        PluginPhase desiredPhase = normalizePhase(desiredPlugin.getPhase());
        int desiredPriority = normalizePriority(desiredPlugin.getPriority());
        for (V1alpha1WasmPlugin existedCr : existedCrs) {
            if (existedCr == null || existedCr.getSpec() == null) {
                return true;
            }
            PluginPhase currentPhase = normalizePhase(existedCr.getSpec().getPhase());
            int currentPriority = normalizePriority(existedCr.getSpec().getPriority());
            if (!Objects.equals(currentPhase, desiredPhase) || currentPriority != desiredPriority) {
                return true;
            }
        }
        return false;
    }

    private WasmPlugin buildUpdateRequest(List<V1alpha1WasmPlugin> existedCrs, WasmPlugin desiredPlugin) {
        if (CollectionUtils.isEmpty(existedCrs)) {
            return desiredPlugin;
        }

        V1alpha1WasmPlugin existedCr = existedCrs.stream().filter(Objects::nonNull).findFirst().orElse(null);
        if (existedCr == null) {
            return desiredPlugin;
        }

        WasmPlugin currentPlugin = kubernetesModelConverter.wasmPluginFromCr(existedCr);
        if (currentPlugin == null) {
            return desiredPlugin;
        }

        currentPlugin.setPhase(desiredPlugin.getPhase());
        currentPlugin.setPriority(desiredPlugin.getPriority());
        return currentPlugin;
    }

    private PluginPhase normalizePhase(String phase) {
        return ObjectUtils.firstNonNull(PluginPhase.fromName(phase), PluginPhase.UNSPECIFIED);
    }

    private int normalizePriority(Integer priority) {
        return priority != null ? priority : 0;
    }

    private String getPluginVersion(V1alpha1WasmPlugin cr) {
        if (cr == null || cr.getMetadata() == null) {
            return null;
        }
        return MapUtils.getString(
            cr.getMetadata().getLabels(),
            KubernetesConstants.Label.WASM_PLUGIN_VERSION_KEY
        );
    }
}
