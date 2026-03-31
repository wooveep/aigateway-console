package com.alibaba.higress.console.controller;

import javax.annotation.Resource;

import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import com.alibaba.higress.console.aop.AllowAnonymous;
import com.alibaba.higress.console.model.aisensitive.AiSensitiveBlockAuditEvent;
import com.alibaba.higress.console.service.AiSensitiveWordService;

@RestController("InternalAiSensitiveAuditController")
@RequestMapping("/v1/internal/ai/sensitive-block-events")
@AllowAnonymous
public class InternalAiSensitiveAuditController {

    private AiSensitiveWordService aiSensitiveWordService;

    @Resource
    public void setAiSensitiveWordService(AiSensitiveWordService aiSensitiveWordService) {
        this.aiSensitiveWordService = aiSensitiveWordService;
    }

    @PostMapping
    public ResponseEntity<String> ingest(@RequestBody AiSensitiveBlockAuditEvent event) {
        aiSensitiveWordService.ingestBlockedEvent(event);
        return ResponseEntity.ok("ok");
    }
}
