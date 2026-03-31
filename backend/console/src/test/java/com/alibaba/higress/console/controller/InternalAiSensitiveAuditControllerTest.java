package com.alibaba.higress.console.controller;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.http.ResponseEntity;

import com.alibaba.higress.console.model.aisensitive.AiSensitiveBlockAuditEvent;
import com.alibaba.higress.console.service.AiSensitiveWordService;

class InternalAiSensitiveAuditControllerTest {

    private AiSensitiveWordService aiSensitiveWordService;
    private InternalAiSensitiveAuditController controller;

    @BeforeEach
    void setUp() {
        aiSensitiveWordService = mock(AiSensitiveWordService.class);
        controller = new InternalAiSensitiveAuditController();
        controller.setAiSensitiveWordService(aiSensitiveWordService);
    }

    @Test
    void ingestShouldDelegateBlockedEventAndReturnOk() {
        AiSensitiveBlockAuditEvent event = AiSensitiveBlockAuditEvent.builder()
            .requestId("req-1")
            .consumerName("consumer-a")
            .matchedRule("南京")
            .build();

        ResponseEntity<String> response = controller.ingest(event);

        assertEquals(200, response.getStatusCodeValue());
        assertEquals("ok", response.getBody());
        verify(aiSensitiveWordService).ingestBlockedEvent(event);
    }
}
