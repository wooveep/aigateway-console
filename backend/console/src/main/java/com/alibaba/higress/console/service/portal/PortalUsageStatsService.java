package com.alibaba.higress.console.service.portal;

import java.net.URI;
import java.util.ArrayList;
import java.util.Comparator;
import java.util.HashMap;
import java.util.List;
import java.util.Locale;
import java.util.Map;

import org.apache.commons.lang3.StringUtils;
import org.apache.http.client.methods.CloseableHttpResponse;
import org.apache.http.client.methods.HttpGet;
import org.apache.http.client.utils.URIBuilder;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClients;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import com.alibaba.higress.console.constant.SystemConfigKey;
import com.alibaba.higress.console.model.portal.PortalUsageStatRecord;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;

import lombok.extern.slf4j.Slf4j;

@Slf4j
@Service
public class PortalUsageStatsService {

    private static final String METRIC_INPUT = "sum by(ai_consumer, ai_model) (increase(route_upstream_model_consumer_metric_input_token[%s]))";
    private static final String METRIC_OUTPUT = "sum by(ai_consumer, ai_model) (increase(route_upstream_model_consumer_metric_output_token[%s]))";
    private static final String METRIC_REQUEST = "sum by(ai_consumer, ai_model) (increase(route_upstream_model_consumer_metric_llm_stream_duration_count[%s]))";

    private final ObjectMapper objectMapper = new ObjectMapper();

    @Value("${" + SystemConfigKey.DASHBOARD_DATASOURCE_PROM_URL_KEY + ":}")
    private String promDatasourceUrl;

    public List<PortalUsageStatRecord> listUsage(Long fromMillis, Long toMillis) {
        if (StringUtils.isBlank(promDatasourceUrl)) {
            throw new IllegalStateException("Prometheus datasource is unavailable for portal usage stats.");
        }

        long now = System.currentTimeMillis();
        long to = toMillis != null && toMillis > 0 ? toMillis : now;
        long from = fromMillis != null && fromMillis > 0 && fromMillis < to ? fromMillis : to - 3600_000L;
        long rangeSeconds = Math.max(60L, (to - from) / 1000L);
        String range = rangeSeconds + "s";

        Map<UsageKey, Long> inputMap = queryMetric(String.format(Locale.ROOT, METRIC_INPUT, range), to);
        Map<UsageKey, Long> outputMap = queryMetric(String.format(Locale.ROOT, METRIC_OUTPUT, range), to);
        Map<UsageKey, Long> requestMap;
        try {
            requestMap = queryMetric(String.format(Locale.ROOT, METRIC_REQUEST, range), to);
        } catch (Exception ex) {
            log.debug("Request metric query failed, falling back to 0 request count.", ex);
            requestMap = new HashMap<>();
        }

        Map<UsageKey, PortalUsageStatRecord> merged = new HashMap<>();
        mergeMetric(merged, inputMap, true, false, false);
        mergeMetric(merged, outputMap, false, true, false);
        mergeMetric(merged, requestMap, false, false, true);

        List<PortalUsageStatRecord> result = new ArrayList<>(merged.values());
        result.sort(Comparator.comparing(PortalUsageStatRecord::getConsumerName)
            .thenComparing(PortalUsageStatRecord::getModelName));
        return result;
    }

    private void mergeMetric(Map<UsageKey, PortalUsageStatRecord> merged, Map<UsageKey, Long> values,
        boolean input, boolean output, boolean request) {
        for (Map.Entry<UsageKey, Long> entry : values.entrySet()) {
            UsageKey key = entry.getKey();
            long value = entry.getValue();
            PortalUsageStatRecord record = merged.computeIfAbsent(key,
                k -> PortalUsageStatRecord.builder().consumerName(k.consumerName).modelName(k.modelName).build());
            if (input) {
                record.setInputTokens(value);
            }
            if (output) {
                record.setOutputTokens(value);
            }
            if (request) {
                record.setRequestCount(value);
            }
            record.setTotalTokens(record.getInputTokens() + record.getOutputTokens());
        }
    }

    private Map<UsageKey, Long> queryMetric(String expression, long queryTimeMillis) {
        try (CloseableHttpClient client = HttpClients.createDefault()) {
            URI uri = new URIBuilder(buildPrometheusApiUrl("/api/v1/query")).addParameter("query", expression)
                .addParameter("time", String.valueOf(queryTimeMillis / 1000L)).build();
            HttpGet request = new HttpGet(uri);
            try (CloseableHttpResponse response = client.execute(request)) {
                if (response.getStatusLine().getStatusCode() / 100 != 2) {
                    throw new IllegalStateException("Prometheus query failed. Status="
                        + response.getStatusLine().getStatusCode());
                }
                JsonNode rootNode = objectMapper.readTree(response.getEntity().getContent());
                if (!"success".equals(rootNode.path("status").asText())) {
                    throw new IllegalStateException("Prometheus query failed: " + rootNode);
                }
                JsonNode resultNode = rootNode.path("data").path("result");
                Map<UsageKey, Long> values = new HashMap<>();
                if (!resultNode.isArray()) {
                    return values;
                }
                for (JsonNode node : resultNode) {
                    String consumer = node.path("metric").path("ai_consumer").asText("");
                    String model = node.path("metric").path("ai_model").asText("unknown");
                    if (StringUtils.isBlank(consumer)) {
                        continue;
                    }
                    JsonNode valueNode = node.path("value");
                    if (!valueNode.isArray() || valueNode.size() < 2) {
                        continue;
                    }
                    long value = parseLongSafely(valueNode.path(1).asText());
                    values.put(new UsageKey(consumer, model), value);
                }
                return values;
            }
        } catch (Exception ex) {
            throw new IllegalStateException("Failed to query Prometheus metric.", ex);
        }
    }

    private long parseLongSafely(String raw) {
        if (StringUtils.isBlank(raw) || "NaN".equalsIgnoreCase(raw)) {
            return 0L;
        }
        try {
            return Math.round(Double.parseDouble(raw));
        } catch (NumberFormatException ex) {
            return 0L;
        }
    }

    private String buildPrometheusApiUrl(String apiPath) {
        String baseUrl = StringUtils.removeEnd(promDatasourceUrl, "/");
        if (apiPath.startsWith("/")) {
            return baseUrl + apiPath;
        }
        return baseUrl + "/" + apiPath;
    }

    private static class UsageKey {

        private final String consumerName;
        private final String modelName;

        private UsageKey(String consumerName, String modelName) {
            this.consumerName = consumerName;
            this.modelName = modelName;
        }

        @Override
        public boolean equals(Object obj) {
            if (this == obj) {
                return true;
            }
            if (obj == null || getClass() != obj.getClass()) {
                return false;
            }
            UsageKey that = (UsageKey) obj;
            return consumerName.equals(that.consumerName) && modelName.equals(that.modelName);
        }

        @Override
        public int hashCode() {
            return java.util.Objects.hash(consumerName, modelName);
        }
    }
}
