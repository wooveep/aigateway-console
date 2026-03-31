package com.alibaba.higress.console.controller;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import java.util.Collections;
import java.util.List;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.http.ResponseEntity;

import com.alibaba.higress.console.controller.dto.Response;
import com.alibaba.higress.console.model.aisensitive.AiSensitiveBlockAudit;
import com.alibaba.higress.console.model.aisensitive.AiSensitiveSystemConfig;
import com.alibaba.higress.console.service.AiSensitiveWordService;

class AiSensitiveWordControllerTest {

    private AiSensitiveWordService aiSensitiveWordService;
    private AiSensitiveWordController controller;

    @BeforeEach
    void setUp() {
        aiSensitiveWordService = mock(AiSensitiveWordService.class);
        controller = new AiSensitiveWordController();
        controller.setAiSensitiveWordService(aiSensitiveWordService);
    }

    @Test
    void listAuditsShouldReturnServiceResult() {
        String startTime = "2026-03-30T00:00:00Z";
        String endTime = "2026-03-31T00:00:00Z";
        List<AiSensitiveBlockAudit> audits = Collections.singletonList(
            AiSensitiveBlockAudit.builder()
                .id(1L)
                .requestId("req-1")
                .consumerName("consumer-a")
                .displayName("Demo User")
                .matchedRule("南京")
                .build()
        );

        when(aiSensitiveWordService.listAudits(
            "consumer-a",
            "Demo User",
            "ai-route-doubao.internal",
            "contains",
            startTime,
            endTime,
            20
        )).thenReturn(audits);

        ResponseEntity<Response<List<AiSensitiveBlockAudit>>> response = controller.listAudits(
            "consumer-a",
            "Demo User",
            "ai-route-doubao.internal",
            "contains",
            startTime,
            endTime,
            20
        );

        assertEquals(200, response.getStatusCodeValue());
        assertEquals(audits, response.getBody().getData());
        verify(aiSensitiveWordService).listAudits(
            "consumer-a",
            "Demo User",
            "ai-route-doubao.internal",
            "contains",
            startTime,
            endTime,
            20
        );
    }

    @Test
    void getSystemConfigShouldReturnServiceResult() {
        AiSensitiveSystemConfig config = AiSensitiveSystemConfig.builder()
            .systemDenyEnabled(Boolean.FALSE)
            .dictionaryText("天安门")
            .build();
        when(aiSensitiveWordService.getSystemConfig()).thenReturn(config);

        ResponseEntity<Response<AiSensitiveSystemConfig>> response = controller.getSystemConfig();

        assertEquals(200, response.getStatusCodeValue());
        assertEquals(config, response.getBody().getData());
        verify(aiSensitiveWordService).getSystemConfig();
    }
}
