package com.alibaba.higress.console.service;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertSame;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import java.util.Arrays;
import java.util.Collections;
import java.util.List;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import com.alibaba.higress.console.model.aisensitive.AiSensitiveBlockAudit;
import com.alibaba.higress.console.model.aisensitive.AiSensitiveBlockAuditEvent;
import com.alibaba.higress.console.model.aisensitive.AiSensitiveMenuState;
import com.alibaba.higress.console.model.aisensitive.AiSensitiveSystemConfig;
import com.alibaba.higress.console.model.User;
import com.alibaba.higress.console.model.portal.PortalUserRecord;
import com.alibaba.higress.console.service.portal.AiSensitiveWordJdbcService;
import com.alibaba.higress.console.service.portal.PortalUserJdbcService;
import com.alibaba.higress.sdk.constant.plugin.BuiltInPluginName;
import com.alibaba.higress.sdk.model.WasmPluginInstance;
import com.alibaba.higress.sdk.model.WasmPluginInstanceScope;
import com.alibaba.higress.sdk.service.WasmPluginInstanceService;

class AiSensitiveWordServiceImplTest {

    private AiSensitiveWordJdbcService aiSensitiveWordJdbcService;
    private PortalUserJdbcService portalUserJdbcService;
    private AiSensitiveWordProjectionService projectionService;
    private WasmPluginInstanceService wasmPluginInstanceService;
    private AiSensitiveWordServiceImpl service;

    @BeforeEach
    void setUp() {
        SessionUserHelper.clearCurrentUser();
        aiSensitiveWordJdbcService = mock(AiSensitiveWordJdbcService.class);
        portalUserJdbcService = mock(PortalUserJdbcService.class);
        projectionService = mock(AiSensitiveWordProjectionService.class);
        wasmPluginInstanceService = mock(WasmPluginInstanceService.class);

        service = new AiSensitiveWordServiceImpl();
        service.setAiSensitiveWordJdbcService(aiSensitiveWordJdbcService);
        service.setPortalUserJdbcService(portalUserJdbcService);
        service.setProjectionService(projectionService);
        service.setWasmPluginInstanceService(wasmPluginInstanceService);
    }

    @Test
    void ingestBlockedEventShouldResolveDisplayNameFromPortalUser() {
        AiSensitiveBlockAuditEvent event = AiSensitiveBlockAuditEvent.builder()
            .consumerName("consumer-a")
            .matchedRule("南京")
            .build();
        PortalUserRecord portalUser = PortalUserRecord.builder()
            .consumerName("consumer-a")
            .displayName("Demo User")
            .build();
        AiSensitiveBlockAudit audit = AiSensitiveBlockAudit.builder()
            .id(1L)
            .consumerName("consumer-a")
            .displayName("Demo User")
            .matchedRule("南京")
            .build();

        when(portalUserJdbcService.queryByConsumerName("consumer-a")).thenReturn(portalUser);
        when(aiSensitiveWordJdbcService.saveAudit(event, "Demo User")).thenReturn(audit);

        AiSensitiveBlockAudit result = service.ingestBlockedEvent(event);

        assertSame(audit, result);
        assertEquals("Demo User", result.getDisplayName());
        verify(portalUserJdbcService).queryByConsumerName("consumer-a");
        verify(aiSensitiveWordJdbcService).saveAudit(event, "Demo User");
    }

    @Test
    void saveSystemConfigShouldUseCurrentUserDisplayName() {
        SessionUserHelper.setCurrentUser(User.builder().name("alice").displayName("Alice").build());
        AiSensitiveSystemConfig config = AiSensitiveSystemConfig.builder()
            .systemDenyEnabled(Boolean.TRUE)
            .dictionaryText("天安门")
            .build();
        AiSensitiveSystemConfig saved = AiSensitiveSystemConfig.builder()
            .systemDenyEnabled(Boolean.TRUE)
            .dictionaryText("天安门")
            .updatedBy("Alice")
            .build();

        when(aiSensitiveWordJdbcService.saveSystemConfig(config, "Alice")).thenReturn(saved);

        AiSensitiveSystemConfig result = service.saveSystemConfig(config);

        assertSame(saved, result);
        verify(aiSensitiveWordJdbcService).saveSystemConfig(config, "Alice");
        verify(projectionService).syncNow();
    }

    @Test
    void getMenuStateShouldCountEnabledRouteInstances() {
        WasmPluginInstance enabledRouteInstance = WasmPluginInstance.builder()
            .enabled(true)
            .targets(Collections.singletonMap(WasmPluginInstanceScope.ROUTE, "ai-route-a"))
            .build();
        WasmPluginInstance disabledRouteInstance = WasmPluginInstance.builder()
            .enabled(false)
            .targets(Collections.singletonMap(WasmPluginInstanceScope.ROUTE, "ai-route-b"))
            .build();
        WasmPluginInstance enabledGlobalInstance = WasmPluginInstance.builder()
            .enabled(true)
            .targets(Collections.singletonMap(WasmPluginInstanceScope.GLOBAL, "global"))
            .build();

        when(wasmPluginInstanceService.list(BuiltInPluginName.AI_DATA_MASKING, false))
            .thenReturn(Arrays.asList(enabledRouteInstance, disabledRouteInstance, enabledGlobalInstance));

        AiSensitiveMenuState result = service.getMenuState();

        assertEquals(true, result.isEnabled());
        assertEquals(1, result.getEnabledRouteCount());
        verify(wasmPluginInstanceService).list(BuiltInPluginName.AI_DATA_MASKING, false);
    }
}
