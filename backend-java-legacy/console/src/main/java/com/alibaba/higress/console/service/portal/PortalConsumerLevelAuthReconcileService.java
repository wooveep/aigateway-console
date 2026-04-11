/*
 * Copyright (c) 2022-2024 Alibaba Group Holding Ltd.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */
package com.alibaba.higress.console.service.portal;

import java.util.Collections;
import java.util.List;
import java.util.Objects;
import java.util.stream.Collectors;

import javax.annotation.Resource;

import org.apache.commons.collections4.CollectionUtils;
import org.apache.commons.lang3.StringUtils;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;

import com.alibaba.higress.sdk.constant.CommonKey;
import com.alibaba.higress.sdk.constant.HigressConstants;
import com.alibaba.higress.sdk.model.PaginatedResult;
import com.alibaba.higress.sdk.model.Route;
import com.alibaba.higress.sdk.model.RouteAuthConfig;
import com.alibaba.higress.sdk.model.ai.AiRoute;
import com.alibaba.higress.sdk.model.mcp.ConsumerAuthInfo;
import com.alibaba.higress.sdk.model.mcp.McpServer;
import com.alibaba.higress.sdk.service.RouteService;
import com.alibaba.higress.sdk.service.ai.AiRouteService;
import com.alibaba.higress.sdk.service.mcp.McpServerService;

import lombok.extern.slf4j.Slf4j;

@Slf4j
@Service
public class PortalConsumerLevelAuthReconcileService {

    @Value("${higress.portal.level-auth.reconcile.enabled:true}")
    private boolean reconcileEnabled;

    @Resource
    private PortalUserJdbcService portalUserJdbcService;

    @Resource
    private PortalConsumerLevelAuthService portalConsumerLevelAuthService;

    @Resource
    private RouteService routeService;

    @Resource
    private AiRouteService aiRouteService;

    @Resource
    private McpServerService mcpServerService;

    @Scheduled(initialDelayString = "${higress.portal.level-auth.reconcile.initial-delay-millis:10000}",
        fixedDelayString = "${higress.portal.level-auth.reconcile.interval-millis:5000}")
    public void scheduledReconcile() {
        reconcileNow("scheduled");
    }

    public synchronized void reconcileNow(String trigger) {
        if (!reconcileEnabled) {
            return;
        }
        if (!portalUserJdbcService.enabled()) {
            return;
        }

        int routeUpdated = reconcileRoutes();
        int aiRouteUpdated = reconcileAiRoutes();
        int mcpUpdated = reconcileMcpServers();
        if (routeUpdated > 0 || aiRouteUpdated > 0 || mcpUpdated > 0) {
            log.info("Reconciled consumer allow-list by user levels. trigger={}, routeUpdated={}, aiRouteUpdated={}, mcpUpdated={}",
                StringUtils.defaultIfBlank(trigger, "manual"), routeUpdated, aiRouteUpdated, mcpUpdated);
        }
    }

    private int reconcileRoutes() {
        PaginatedResult<Route> paginatedResult;
        try {
            paginatedResult = routeService.list(null);
        } catch (Exception ex) {
            log.warn("Failed to list routes when reconciling consumer allow-list by levels.", ex);
            return 0;
        }
        if (paginatedResult == null || CollectionUtils.isEmpty(paginatedResult.getData())) {
            return 0;
        }

        int updated = 0;
        for (Route route : paginatedResult.getData()) {
            if (!shouldReconcileRoute(route)) {
                continue;
            }
            RouteAuthConfig authConfig = route.getAuthConfig();
            List<String> beforeLevels = normalizeLevels(authConfig.getAllowedConsumerLevels());
            List<String> beforeConsumers = normalizeConsumers(authConfig.getAllowedConsumers());
            try {
                portalConsumerLevelAuthService.resolveRouteAuthConfig(authConfig);
            } catch (Exception ex) {
                log.warn("Failed to resolve route {} auth config by levels.", route.getName(), ex);
                continue;
            }
            List<String> afterLevels = normalizeLevels(authConfig.getAllowedConsumerLevels());
            List<String> afterConsumers = normalizeConsumers(authConfig.getAllowedConsumers());
            if (Objects.equals(beforeLevels, afterLevels) && Objects.equals(beforeConsumers, afterConsumers)) {
                continue;
            }
            try {
                routeService.update(route);
                updated++;
            } catch (Exception ex) {
                log.warn("Failed to reconcile route {} consumer allow-list by levels.", route.getName(), ex);
            }
        }
        return updated;
    }

    private int reconcileAiRoutes() {
        PaginatedResult<AiRoute> paginatedResult;
        try {
            paginatedResult = aiRouteService.list(null);
        } catch (Exception ex) {
            log.warn("Failed to list AI routes when reconciling consumer allow-list by levels.", ex);
            return 0;
        }
        if (paginatedResult == null || CollectionUtils.isEmpty(paginatedResult.getData())) {
            return 0;
        }

        int updated = 0;
        for (AiRoute route : paginatedResult.getData()) {
            if (route == null || route.getAuthConfig() == null) {
                continue;
            }
            RouteAuthConfig authConfig = route.getAuthConfig();
            List<String> beforeLevels = normalizeLevels(authConfig.getAllowedConsumerLevels());
            List<String> beforeConsumers = normalizeConsumers(authConfig.getAllowedConsumers());
            try {
                portalConsumerLevelAuthService.resolveAiRouteAuthConfig(route);
            } catch (Exception ex) {
                log.warn("Failed to resolve AI route {} auth config by levels.", route.getName(), ex);
                continue;
            }
            List<String> afterLevels = normalizeLevels(authConfig.getAllowedConsumerLevels());
            List<String> afterConsumers = normalizeConsumers(authConfig.getAllowedConsumers());
            if (Objects.equals(beforeLevels, afterLevels) && Objects.equals(beforeConsumers, afterConsumers)) {
                continue;
            }
            try {
                aiRouteService.update(route);
                updated++;
            } catch (Exception ex) {
                log.warn("Failed to reconcile AI route {} consumer allow-list by levels.", route.getName(), ex);
            }
        }
        return updated;
    }

    private int reconcileMcpServers() {
        PaginatedResult<McpServer> paginatedResult;
        try {
            paginatedResult = mcpServerService.list(null);
        } catch (Exception ex) {
            log.warn("Failed to list MCP servers when reconciling consumer allow-list by levels.", ex);
            return 0;
        }
        if (paginatedResult == null || CollectionUtils.isEmpty(paginatedResult.getData())) {
            return 0;
        }

        int updated = 0;
        for (McpServer item : paginatedResult.getData()) {
            if (item == null || StringUtils.isBlank(item.getName())) {
                continue;
            }
            McpServer detail;
            try {
                detail = mcpServerService.query(item.getName());
            } catch (Exception ex) {
                log.warn("Failed to query MCP server {} when reconciling consumer allow-list by levels.", item.getName(),
                    ex);
                continue;
            }
            ConsumerAuthInfo authInfo = detail == null ? null : detail.getConsumerAuthInfo();
            if (authInfo == null) {
                continue;
            }
            List<String> beforeLevels = normalizeLevels(authInfo.getAllowedConsumerLevels());
            List<String> beforeConsumers = normalizeConsumers(authInfo.getAllowedConsumers());
            try {
                portalConsumerLevelAuthService.resolveMcpConsumerAuth(detail);
            } catch (Exception ex) {
                log.warn("Failed to resolve MCP server {} auth config by levels.", detail.getName(), ex);
                continue;
            }
            List<String> afterLevels = normalizeLevels(authInfo.getAllowedConsumerLevels());
            List<String> afterConsumers = normalizeConsumers(authInfo.getAllowedConsumers());
            if (Objects.equals(beforeLevels, afterLevels) && Objects.equals(beforeConsumers, afterConsumers)) {
                continue;
            }
            try {
                mcpServerService.addOrUpdateWithAuthorization(detail);
                updated++;
            } catch (Exception ex) {
                log.warn("Failed to reconcile MCP server {} consumer allow-list by levels.", detail.getName(), ex);
            }
        }
        return updated;
    }

    private boolean shouldReconcileRoute(Route route) {
        if (route == null || route.getAuthConfig() == null || StringUtils.isBlank(route.getName())) {
            return false;
        }
        String routeName = route.getName();
        if (StringUtils.startsWith(routeName, CommonKey.MCP_SERVER_ROUTE_PREFIX)) {
            return false;
        }
        if (StringUtils.startsWith(routeName, CommonKey.AI_ROUTE_PREFIX)
            && StringUtils.endsWith(routeName, HigressConstants.INTERNAL_RESOURCE_NAME_SUFFIX)) {
            return false;
        }
        return hasReconcileTarget(route.getAuthConfig().getAllowedConsumerLevels(), route.getAuthConfig().getAllowedConsumers());
    }

    private boolean hasReconcileTarget(List<String> levels, List<String> consumers) {
        return CollectionUtils.isNotEmpty(normalizeLevels(levels)) || CollectionUtils.isNotEmpty(normalizeConsumers(consumers));
    }

    private List<String> normalizeLevels(List<String> levels) {
        if (CollectionUtils.isEmpty(levels)) {
            return Collections.emptyList();
        }
        try {
            return RouteAuthConfig.normalizeAllowedConsumerLevels(levels);
        } catch (Exception ex) {
            log.warn("Invalid allowed consumer levels found during reconcile: {}", levels, ex);
            return Collections.emptyList();
        }
    }

    private List<String> normalizeConsumers(List<String> consumers) {
        if (CollectionUtils.isEmpty(consumers)) {
            return Collections.emptyList();
        }
        return consumers.stream().map(StringUtils::trimToNull).filter(Objects::nonNull).distinct().sorted().collect(
            Collectors.toList());
    }
}
