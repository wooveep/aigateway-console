package com.alibaba.higress.console.service;

import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.time.Instant;
import java.time.ZoneId;
import java.time.ZonedDateTime;
import java.util.ArrayList;
import java.util.Collections;
import java.util.Comparator;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Objects;
import java.util.UUID;
import java.util.function.Function;
import java.util.stream.Collectors;

import javax.annotation.Resource;

import org.apache.commons.collections4.CollectionUtils;
import org.apache.commons.collections4.MapUtils;
import org.apache.commons.lang3.ObjectUtils;
import org.apache.commons.lang3.StringUtils;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.scheduling.support.CronExpression;
import org.springframework.stereotype.Service;

import com.alibaba.higress.console.model.aiquota.AiQuotaConsumerQuota;
import com.alibaba.higress.console.model.aiquota.AiQuotaMenuState;
import com.alibaba.higress.console.model.aiquota.AiQuotaRouteSummary;
import com.alibaba.higress.console.model.aiquota.AiQuotaScheduleRule;
import com.alibaba.higress.console.model.aiquota.AiQuotaScheduleRuleRequest;
import com.alibaba.higress.sdk.constant.CommonKey;
import com.alibaba.higress.sdk.constant.HigressConstants;
import com.alibaba.higress.sdk.constant.KubernetesConstants;
import com.alibaba.higress.sdk.constant.plugin.BuiltInPluginName;
import com.alibaba.higress.sdk.exception.BusinessException;
import com.alibaba.higress.sdk.model.PaginatedResult;
import com.alibaba.higress.sdk.model.WasmPluginInstance;
import com.alibaba.higress.sdk.model.WasmPluginInstanceScope;
import com.alibaba.higress.sdk.model.ai.AiRoute;
import com.alibaba.higress.sdk.model.consumer.Consumer;
import com.alibaba.higress.sdk.service.WasmPluginInstanceService;
import com.alibaba.higress.sdk.service.ai.AiRouteService;
import com.alibaba.higress.sdk.service.consumer.ConsumerService;
import com.alibaba.higress.sdk.service.kubernetes.KubernetesClientService;
import com.alibaba.higress.sdk.service.kubernetes.KubernetesUtil;
import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;

import io.kubernetes.client.openapi.ApiException;
import io.kubernetes.client.openapi.models.V1ConfigMap;
import io.kubernetes.client.openapi.models.V1ObjectMeta;
import lombok.Data;
import lombok.extern.slf4j.Slf4j;
import redis.clients.jedis.Jedis;

@Slf4j
@Service
public class AiQuotaServiceImpl implements AiQuotaService {

    private static final String SCHEDULE_CONFIG_MAP_PREFIX = "aiq";
    private static final String SCHEDULE_CONFIG_MAP_RULES_KEY = "rules";
    private static final String SCHEDULE_CONFIG_MAP_ROUTE_KEY = "routeName";
    private static final String SCHEDULE_BIZ_TYPE = "ai-quota-schedule";
    private static final String ACTION_REFRESH = "REFRESH";
    private static final String ACTION_DELTA = "DELTA";
    private static final int DEFAULT_REDIS_PORT = 6379;
    private static final int DEFAULT_STATIC_SERVICE_REDIS_PORT = 80;
    private static final int DEFAULT_TIMEOUT_MILLIS = 1000;
    private static final String DEFAULT_ADMIN_PATH = "/quota";
    private static final String DEFAULT_REDIS_KEY_PREFIX = "chat_quota:";
    private static final int SCHEDULE_NAME_ROUTE_PART_MAX_LENGTH = 48;

    private static final TypeReference<List<AiQuotaScheduleRule>> RULE_LIST_TYPE =
        new TypeReference<List<AiQuotaScheduleRule>>() {
        };

    private AiRouteService aiRouteService;
    private WasmPluginInstanceService wasmPluginInstanceService;
    private ConsumerService consumerService;
    private KubernetesClientService kubernetesClientService;
    private ObjectMapper objectMapper;

    @Resource
    public void setAiRouteService(AiRouteService aiRouteService) {
        this.aiRouteService = aiRouteService;
    }

    @Resource
    public void setWasmPluginInstanceService(WasmPluginInstanceService wasmPluginInstanceService) {
        this.wasmPluginInstanceService = wasmPluginInstanceService;
    }

    @Resource
    public void setConsumerService(ConsumerService consumerService) {
        this.consumerService = consumerService;
    }

    @Resource
    public void setKubernetesClientService(KubernetesClientService kubernetesClientService) {
        this.kubernetesClientService = kubernetesClientService;
    }

    @Resource
    public void setObjectMapper(ObjectMapper objectMapper) {
        this.objectMapper = objectMapper;
    }

    @Override
    public AiQuotaMenuState getMenuState() {
        int enabledRouteCount = listEnabledRouteContexts().size();
        return AiQuotaMenuState.builder().enabled(enabledRouteCount > 0).enabledRouteCount(enabledRouteCount).build();
    }

    @Override
    public List<AiQuotaRouteSummary> listEnabledRoutes() {
        Map<String, Integer> scheduleCountMap = buildScheduleCountMap();
        return listEnabledRouteContexts().stream()
            .map(ctx -> AiQuotaRouteSummary.builder()
                .routeName(ctx.getAiRoute().getName())
                .domains(ctx.getAiRoute().getDomains())
                .path(ctx.getAiRoute().getPathPredicate() != null ? ctx.getAiRoute().getPathPredicate().getMatchValue()
                    : null)
                .redisKeyPrefix(ctx.getQuotaRouteConfig().getRedisKeyPrefix())
                .adminConsumer(ctx.getQuotaRouteConfig().getAdminConsumer())
                .adminPath(ctx.getQuotaRouteConfig().getAdminPath())
                .scheduleRuleCount(scheduleCountMap.getOrDefault(ctx.getAiRoute().getName(), 0))
                .build())
            .sorted(Comparator.comparing(AiQuotaRouteSummary::getRouteName))
            .collect(Collectors.toList());
    }

    @Override
    public List<AiQuotaConsumerQuota> listConsumerQuotas(String routeName) {
        AiQuotaRouteContext routeContext = requireRouteContext(routeName);
        List<Consumer> consumers = listConsumers();
        if (CollectionUtils.isEmpty(consumers)) {
            return Collections.emptyList();
        }
        return withRedis(routeContext.getQuotaRouteConfig().getRedisConfig(), jedis -> consumers.stream()
            .sorted(Comparator.comparing(Consumer::getName))
            .map(consumer -> AiQuotaConsumerQuota.builder()
                .consumerName(consumer.getName())
                .quota(parseQuota(jedis.get(buildQuotaKey(routeContext.getQuotaRouteConfig(), consumer.getName()))))
                .build())
            .collect(Collectors.toList()));
    }

    @Override
    public AiQuotaConsumerQuota refreshQuota(String routeName, String consumerName, int quota) {
        validateConsumerExists(consumerName);
        AiQuotaRouteContext routeContext = requireRouteContext(routeName);
        withRedis(routeContext.getQuotaRouteConfig().getRedisConfig(), jedis -> {
            jedis.set(buildQuotaKey(routeContext.getQuotaRouteConfig(), consumerName), Integer.toString(quota));
            return null;
        });
        return AiQuotaConsumerQuota.builder().consumerName(consumerName).quota(quota).build();
    }

    @Override
    public AiQuotaConsumerQuota deltaQuota(String routeName, String consumerName, int delta) {
        validateConsumerExists(consumerName);
        AiQuotaRouteContext routeContext = requireRouteContext(routeName);
        int quota = withRedis(routeContext.getQuotaRouteConfig().getRedisConfig(), jedis -> {
            long result = jedis.incrBy(buildQuotaKey(routeContext.getQuotaRouteConfig(), consumerName), delta);
            return Math.toIntExact(result);
        });
        return AiQuotaConsumerQuota.builder().consumerName(consumerName).quota(quota).build();
    }

    @Override
    public List<AiQuotaScheduleRule> listScheduleRules(String routeName, String consumerName) {
        requireRouteContext(routeName);
        ScheduleStore store = readScheduleStore(routeName);
        return store.getRules().stream()
            .filter(rule -> StringUtils.isBlank(consumerName) || StringUtils.equals(consumerName, rule.getConsumerName()))
            .sorted(Comparator.comparing(AiQuotaScheduleRule::getConsumerName)
                .thenComparing(AiQuotaScheduleRule::getCreatedAt, Comparator.nullsLast(Long::compareTo)))
            .collect(Collectors.toList());
    }

    @Override
    public AiQuotaScheduleRule saveScheduleRule(String routeName, AiQuotaScheduleRuleRequest request) {
        requireRouteContext(routeName);
        validateScheduleRequest(request);
        validateConsumerExists(request.getConsumerName());

        ScheduleStore store = readScheduleStore(routeName);
        List<AiQuotaScheduleRule> rules = new ArrayList<>(store.getRules());
        long now = System.currentTimeMillis();

        AiQuotaScheduleRule target = null;
        if (StringUtils.isNotBlank(request.getId())) {
            target = rules.stream().filter(rule -> StringUtils.equals(rule.getId(), request.getId())).findFirst()
                .orElseThrow(() -> new IllegalArgumentException("Unknown schedule rule: " + request.getId()));
        }

        if (target == null) {
            target = new AiQuotaScheduleRule();
            target.setId(UUID.randomUUID().toString());
            target.setCreatedAt(now);
            rules.add(target);
        }

        target.setConsumerName(request.getConsumerName());
        target.setAction(StringUtils.upperCase(request.getAction()));
        target.setCron(request.getCron().trim());
        target.setValue(request.getValue());
        target.setEnabled(ObjectUtils.defaultIfNull(request.getEnabled(), Boolean.TRUE));
        target.setUpdatedAt(now);

        store.setRules(rules);
        saveScheduleStore(store);
        return target;
    }

    @Override
    public void deleteScheduleRule(String routeName, String ruleId) {
        requireRouteContext(routeName);
        if (StringUtils.isBlank(ruleId)) {
            throw new IllegalArgumentException("ruleId cannot be empty.");
        }
        ScheduleStore store = readScheduleStore(routeName);
        if (CollectionUtils.isEmpty(store.getRules())) {
            return;
        }
        int originalSize = store.getRules().size();
        store.getRules().removeIf(rule -> StringUtils.equals(ruleId, rule.getId()));
        if (store.getRules().size() != originalSize) {
            saveScheduleStore(store);
        }
    }

    @Scheduled(initialDelay = 60000L, fixedDelay = 30000L)
    public void executeScheduledRules() {
        List<ScheduleStore> stores = listAllScheduleStores();
        if (CollectionUtils.isEmpty(stores)) {
            return;
        }
        long now = System.currentTimeMillis();
        for (ScheduleStore store : stores) {
            boolean changed = false;
            for (AiQuotaScheduleRule rule : store.getRules()) {
                if (!Boolean.TRUE.equals(rule.getEnabled()) || !shouldRun(rule, now)) {
                    continue;
                }
                try {
                    if (ACTION_REFRESH.equalsIgnoreCase(rule.getAction())) {
                        refreshQuota(store.getRouteName(), rule.getConsumerName(), rule.getValue());
                    } else {
                        deltaQuota(store.getRouteName(), rule.getConsumerName(), rule.getValue());
                    }
                    rule.setLastAppliedAt(now);
                    rule.setLastError(null);
                } catch (Exception ex) {
                    rule.setLastError(StringUtils.abbreviate(ex.getMessage(), 512));
                    log.warn("Failed to execute ai quota schedule. route={}, consumer={}, rule={}",
                        store.getRouteName(), rule.getConsumerName(), rule.getId(), ex);
                }
                changed = true;
            }
            if (changed) {
                try {
                    saveScheduleStore(store);
                } catch (Exception ex) {
                    log.warn("Failed to persist ai quota schedule execution result for route={}",
                        store.getRouteName(), ex);
                }
            }
        }
    }

    private Map<String, Integer> buildScheduleCountMap() {
        Map<String, Integer> result = new HashMap<>();
        for (ScheduleStore store : listAllScheduleStores()) {
            result.put(store.getRouteName(), store.getRules().size());
        }
        return result;
    }

    private List<AiQuotaRouteContext> listEnabledRouteContexts() {
        List<WasmPluginInstance> instances = wasmPluginInstanceService.list(BuiltInPluginName.AI_QUOTA, false);
        if (CollectionUtils.isEmpty(instances)) {
            return Collections.emptyList();
        }

        Map<String, AiRoute> aiRouteMap = listAiRoutes().stream()
            .filter(route -> StringUtils.isNotBlank(route.getName()))
            .collect(Collectors.toMap(AiRoute::getName, Function.identity(), (first, second) -> first));

        List<AiQuotaRouteContext> results = new ArrayList<>();
        for (WasmPluginInstance instance : instances) {
            if (!Boolean.TRUE.equals(instance.getEnabled()) || !instance.hasScopedTarget(WasmPluginInstanceScope.ROUTE)) {
                continue;
            }
            String routeResourceName = instance.getTargets().get(WasmPluginInstanceScope.ROUTE);
            String aiRouteName = toAiRouteName(routeResourceName);
            if (StringUtils.isBlank(aiRouteName)) {
                continue;
            }
            AiRoute aiRoute = aiRouteMap.get(aiRouteName);
            if (aiRoute == null) {
                continue;
            }
            AiQuotaRouteConfig quotaRouteConfig = parseQuotaRouteConfig(instance.getConfigurations());
            results.add(new AiQuotaRouteContext(aiRoute, routeResourceName, quotaRouteConfig));
        }
        return results;
    }

    private AiQuotaRouteContext requireRouteContext(String routeName) {
        if (StringUtils.isBlank(routeName)) {
            throw new IllegalArgumentException("routeName cannot be empty.");
        }
        return listEnabledRouteContexts().stream().filter(ctx -> StringUtils.equals(routeName, ctx.getAiRoute().getName()))
            .findFirst().orElseThrow(() -> new IllegalArgumentException("ai-quota is not enabled on route: " + routeName));
    }

    private List<AiRoute> listAiRoutes() {
        PaginatedResult<AiRoute> paginatedResult = aiRouteService.list(null);
        if (paginatedResult == null || CollectionUtils.isEmpty(paginatedResult.getData())) {
            return Collections.emptyList();
        }
        return paginatedResult.getData();
    }

    private AiQuotaRouteConfig parseQuotaRouteConfig(Map<String, Object> configurations) {
        if (MapUtils.isEmpty(configurations)) {
            throw new IllegalArgumentException("ai-quota configuration cannot be empty.");
        }
        String redisKeyPrefix = ObjectUtils.defaultIfNull(MapUtils.getString(configurations, "redis_key_prefix"),
            DEFAULT_REDIS_KEY_PREFIX);
        String adminConsumer = MapUtils.getString(configurations, "admin_consumer");
        String adminPath = ObjectUtils.defaultIfNull(MapUtils.getString(configurations, "admin_path"), DEFAULT_ADMIN_PATH);
        if (StringUtils.isBlank(adminConsumer)) {
            throw new IllegalArgumentException("ai-quota admin_consumer cannot be empty.");
        }

        Object redisObj = configurations.get("redis");
        if (!(redisObj instanceof Map)) {
            throw new IllegalArgumentException("ai-quota redis configuration cannot be empty.");
        }
        @SuppressWarnings("unchecked")
        Map<String, Object> redisMap = (Map<String, Object>) redisObj;

        String serviceName = MapUtils.getString(redisMap, "service_name");
        if (StringUtils.isBlank(serviceName)) {
            throw new IllegalArgumentException("ai-quota redis.service_name cannot be empty.");
        }
        Integer servicePort = toInteger(redisMap.get("service_port"));
        if (servicePort == null || servicePort <= 0) {
            servicePort = serviceName.endsWith(".static") ? DEFAULT_STATIC_SERVICE_REDIS_PORT : DEFAULT_REDIS_PORT;
        }
        Integer timeout = toInteger(redisMap.get("timeout"));
        if (timeout == null || timeout <= 0) {
            timeout = DEFAULT_TIMEOUT_MILLIS;
        }
        Integer database = ObjectUtils.defaultIfNull(toInteger(redisMap.get("database")), 0);

        RedisConnectionConfig redisConfig = new RedisConnectionConfig();
        redisConfig.setServiceName(serviceName);
        redisConfig.setServicePort(servicePort);
        redisConfig.setUsername(MapUtils.getString(redisMap, "username"));
        redisConfig.setPassword(MapUtils.getString(redisMap, "password"));
        redisConfig.setTimeout(timeout);
        redisConfig.setDatabase(database);

        AiQuotaRouteConfig routeConfig = new AiQuotaRouteConfig();
        routeConfig.setRedisKeyPrefix(redisKeyPrefix);
        routeConfig.setAdminConsumer(adminConsumer);
        routeConfig.setAdminPath(adminPath);
        routeConfig.setRedisConfig(redisConfig);
        return routeConfig;
    }

    private <T> T withRedis(RedisConnectionConfig config, Function<Jedis, T> action) {
        try (Jedis jedis = new Jedis(config.getServiceName(), config.getServicePort(), config.getTimeout(),
            config.getTimeout())) {
            if (StringUtils.isNotBlank(config.getUsername())) {
                jedis.auth(config.getUsername(), StringUtils.defaultString(config.getPassword()));
            } else if (StringUtils.isNotBlank(config.getPassword())) {
                jedis.auth(config.getPassword());
            }
            if (config.getDatabase() != null && config.getDatabase() > 0) {
                jedis.select(config.getDatabase());
            }
            return action.apply(jedis);
        } catch (Exception ex) {
            throw new BusinessException("Failed to access ai-quota Redis: " + ex.getMessage(), ex);
        }
    }

    private void validateConsumerExists(String consumerName) {
        if (StringUtils.isBlank(consumerName)) {
            throw new IllegalArgumentException("consumerName cannot be empty.");
        }
        Consumer consumer = consumerService.query(consumerName);
        if (consumer == null) {
            throw new IllegalArgumentException("Unknown consumer: " + consumerName);
        }
    }

    private List<Consumer> listConsumers() {
        PaginatedResult<Consumer> paginatedResult = consumerService.list(null);
        return paginatedResult != null && paginatedResult.getData() != null ? paginatedResult.getData()
            : Collections.emptyList();
    }

    private void validateScheduleRequest(AiQuotaScheduleRuleRequest request) {
        if (request == null) {
            throw new IllegalArgumentException("schedule rule request cannot be null.");
        }
        if (StringUtils.isBlank(request.getConsumerName())) {
            throw new IllegalArgumentException("consumerName cannot be empty.");
        }
        if (StringUtils.isBlank(request.getAction())) {
            throw new IllegalArgumentException("action cannot be empty.");
        }
        String action = StringUtils.upperCase(request.getAction());
        if (!ACTION_REFRESH.equals(action) && !ACTION_DELTA.equals(action)) {
            throw new IllegalArgumentException("action must be REFRESH or DELTA.");
        }
        if (StringUtils.isBlank(request.getCron())) {
            throw new IllegalArgumentException("cron cannot be empty.");
        }
        try {
            CronExpression.parse(request.getCron().trim());
        } catch (IllegalArgumentException ex) {
            throw new IllegalArgumentException("invalid cron expression: " + request.getCron(), ex);
        }
        if (request.getValue() == null) {
            throw new IllegalArgumentException("value cannot be null.");
        }
    }

    private boolean shouldRun(AiQuotaScheduleRule rule, long now) {
        if (rule == null || StringUtils.isBlank(rule.getCron()) || StringUtils.isBlank(rule.getAction())
            || rule.getValue() == null) {
            return false;
        }
        CronExpression cronExpression;
        try {
            cronExpression = CronExpression.parse(rule.getCron());
        } catch (IllegalArgumentException ex) {
            return false;
        }
        long baseMillis = ObjectUtils.firstNonNull(rule.getLastAppliedAt(), rule.getCreatedAt(), now);
        ZonedDateTime baseTime = Instant.ofEpochMilli(baseMillis).atZone(ZoneId.systemDefault());
        ZonedDateTime nowTime = Instant.ofEpochMilli(now).atZone(ZoneId.systemDefault());
        ZonedDateTime nextTime = cronExpression.next(baseTime);
        return nextTime != null && !nextTime.isAfter(nowTime);
    }

    private ScheduleStore readScheduleStore(String routeName) {
        String configMapName = buildScheduleConfigMapName(routeName);
        V1ConfigMap configMap;
        try {
            configMap = kubernetesClientService.readConfigMap(configMapName);
        } catch (ApiException ex) {
            throw new BusinessException("Failed to read ai quota schedule ConfigMap: " + configMapName, ex);
        }
        if (configMap == null || MapUtils.isEmpty(configMap.getData())) {
            return new ScheduleStore(routeName, new ArrayList<>());
        }
        Map<String, String> data = configMap.getData();
        String storedRouteName = ObjectUtils.defaultIfNull(data.get(SCHEDULE_CONFIG_MAP_ROUTE_KEY), routeName);
        String rawRules = data.get(SCHEDULE_CONFIG_MAP_RULES_KEY);
        if (StringUtils.isBlank(rawRules)) {
            return new ScheduleStore(storedRouteName, new ArrayList<>());
        }
        try {
            List<AiQuotaScheduleRule> rules = objectMapper.readValue(rawRules, RULE_LIST_TYPE);
            return new ScheduleStore(storedRouteName,
                rules != null ? new ArrayList<>(rules) : new ArrayList<>());
        } catch (Exception ex) {
            throw new BusinessException("Failed to parse ai quota schedule ConfigMap: " + configMapName, ex);
        }
    }

    private List<ScheduleStore> listAllScheduleStores() {
        List<V1ConfigMap> configMaps;
        try {
            configMaps = kubernetesClientService.listConfigMap(Collections.singletonMap(
                KubernetesConstants.Label.RESOURCE_BIZ_TYPE_KEY, SCHEDULE_BIZ_TYPE));
        } catch (ApiException ex) {
            throw new BusinessException("Failed to list ai quota schedule ConfigMaps.", ex);
        }
        if (CollectionUtils.isEmpty(configMaps)) {
            return Collections.emptyList();
        }

        List<ScheduleStore> stores = new ArrayList<>();
        for (V1ConfigMap configMap : configMaps) {
            try {
                String routeName = null;
                if (configMap.getData() != null) {
                    routeName = configMap.getData().get(SCHEDULE_CONFIG_MAP_ROUTE_KEY);
                }
                if (StringUtils.isBlank(routeName) && configMap.getMetadata() != null) {
                    routeName = KubernetesUtil.getAnnotation(configMap.getMetadata(), SCHEDULE_CONFIG_MAP_ROUTE_KEY);
                }
                if (StringUtils.isBlank(routeName)) {
                    continue;
                }
                stores.add(readScheduleStore(routeName));
            } catch (Exception ex) {
                log.warn("Failed to load ai quota schedule ConfigMap: {}", KubernetesUtil.getObjectName(configMap), ex);
            }
        }
        return stores;
    }

    private void saveScheduleStore(ScheduleStore store) {
        String configMapName = buildScheduleConfigMapName(store.getRouteName());
        if (CollectionUtils.isEmpty(store.getRules())) {
            try {
                kubernetesClientService.deleteConfigMap(configMapName);
            } catch (ApiException ex) {
                throw new BusinessException("Failed to delete ai quota schedule ConfigMap: " + configMapName, ex);
            }
            return;
        }

        Map<String, String> data = new HashMap<>();
        data.put(SCHEDULE_CONFIG_MAP_ROUTE_KEY, store.getRouteName());
        try {
            data.put(SCHEDULE_CONFIG_MAP_RULES_KEY, objectMapper.writeValueAsString(store.getRules()));
        } catch (Exception ex) {
            throw new BusinessException("Failed to serialize ai quota schedule rules.", ex);
        }

        V1ConfigMap configMap = new V1ConfigMap();
        V1ObjectMeta metadata = new V1ObjectMeta();
        metadata.setName(configMapName);
        KubernetesUtil.setLabel(metadata, KubernetesConstants.Label.RESOURCE_BIZ_TYPE_KEY, SCHEDULE_BIZ_TYPE);
        KubernetesUtil.setAnnotation(metadata, SCHEDULE_CONFIG_MAP_ROUTE_KEY, store.getRouteName());
        configMap.setMetadata(metadata);
        configMap.setData(data);

        try {
            V1ConfigMap existed = kubernetesClientService.readConfigMap(configMapName);
            if (existed == null) {
                kubernetesClientService.createConfigMap(configMap);
            } else {
                metadata.setResourceVersion(existed.getMetadata().getResourceVersion());
                kubernetesClientService.replaceConfigMap(configMap);
            }
        } catch (ApiException ex) {
            throw new BusinessException("Failed to save ai quota schedule ConfigMap: " + configMapName, ex);
        }
    }

    private String buildScheduleConfigMapName(String routeName) {
        String normalizedRouteName = StringUtils.lowerCase(routeName);
        String routePart = StringUtils.left(normalizedRouteName, SCHEDULE_NAME_ROUTE_PART_MAX_LENGTH);
        return String.format("%s-%s-%s", SCHEDULE_CONFIG_MAP_PREFIX, routePart, shortHash(normalizedRouteName));
    }

    private String shortHash(String text) {
        try {
            MessageDigest digest = MessageDigest.getInstance("SHA-256");
            byte[] hashed = digest.digest(text.getBytes(StandardCharsets.UTF_8));
            StringBuilder builder = new StringBuilder();
            for (int i = 0; i < 4; i++) {
                builder.append(String.format("%02x", hashed[i]));
            }
            return builder.toString();
        } catch (Exception ex) {
            throw new IllegalStateException("Failed to hash ai quota schedule key.", ex);
        }
    }

    private String buildQuotaKey(AiQuotaRouteConfig routeConfig, String consumerName) {
        return routeConfig.getRedisKeyPrefix() + consumerName;
    }

    private int parseQuota(String rawQuota) {
        if (StringUtils.isBlank(rawQuota)) {
            return 0;
        }
        try {
            return Integer.parseInt(rawQuota.trim());
        } catch (NumberFormatException ex) {
            return 0;
        }
    }

    private Integer toInteger(Object value) {
        if (value == null) {
            return null;
        }
        if (value instanceof Integer) {
            return (Integer) value;
        }
        if (value instanceof Long) {
            return ((Long) value).intValue();
        }
        if (value instanceof Double) {
            return ((Double) value).intValue();
        }
        if (value instanceof String) {
            try {
                return Integer.parseInt(((String) value).trim());
            } catch (NumberFormatException ex) {
                return null;
            }
        }
        return null;
    }

    private String toAiRouteName(String routeResourceName) {
        String expectedPrefix = CommonKey.AI_ROUTE_PREFIX;
        String expectedSuffix = HigressConstants.INTERNAL_RESOURCE_NAME_SUFFIX;
        if (StringUtils.isBlank(routeResourceName) || !routeResourceName.startsWith(expectedPrefix)
            || !routeResourceName.endsWith(expectedSuffix)) {
            return null;
        }
        String routeName = routeResourceName.substring(expectedPrefix.length(),
            routeResourceName.length() - expectedSuffix.length());
        if (routeName.endsWith(HigressConstants.FALLBACK_ROUTE_NAME_SUFFIX)) {
            return null;
        }
        return routeName;
    }

    @Data
    private static class AiQuotaRouteContext {
        private final AiRoute aiRoute;
        private final String routeResourceName;
        private final AiQuotaRouteConfig quotaRouteConfig;
    }

    @Data
    private static class AiQuotaRouteConfig {
        private String redisKeyPrefix;
        private String adminConsumer;
        private String adminPath;
        private RedisConnectionConfig redisConfig;
    }

    @Data
    private static class RedisConnectionConfig {
        private String serviceName;
        private Integer servicePort;
        private String username;
        private String password;
        private Integer timeout;
        private Integer database;
    }

    @Data
    private static class ScheduleStore {
        private final String routeName;
        private List<AiQuotaScheduleRule> rules;

        private ScheduleStore(String routeName, List<AiQuotaScheduleRule> rules) {
            this.routeName = routeName;
            this.rules = rules;
        }
    }
}
