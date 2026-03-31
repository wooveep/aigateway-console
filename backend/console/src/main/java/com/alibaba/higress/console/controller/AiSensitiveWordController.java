package com.alibaba.higress.console.controller;

import java.util.List;

import javax.annotation.Resource;

import org.springframework.http.ResponseEntity;
import org.springframework.validation.annotation.Validated;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import com.alibaba.higress.console.controller.dto.Response;
import com.alibaba.higress.console.model.aisensitive.AiSensitiveBlockAudit;
import com.alibaba.higress.console.model.aisensitive.AiSensitiveDetectRule;
import com.alibaba.higress.console.model.aisensitive.AiSensitiveReplaceRule;
import com.alibaba.higress.console.model.aisensitive.AiSensitiveStatus;
import com.alibaba.higress.console.model.aisensitive.AiSensitiveSystemConfig;
import com.alibaba.higress.console.service.AiSensitiveWordService;

import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;

@RestController("AiSensitiveWordController")
@RequestMapping("/v1/ai/sensitive-words")
@Validated
@Tag(name = "AI Sensitive Word APIs")
public class AiSensitiveWordController {

    private AiSensitiveWordService aiSensitiveWordService;

    @Resource
    public void setAiSensitiveWordService(AiSensitiveWordService aiSensitiveWordService) {
        this.aiSensitiveWordService = aiSensitiveWordService;
    }

    @GetMapping("/detect-rules")
    @Operation(summary = "List AI sensitive detect rules")
    public ResponseEntity<Response<List<AiSensitiveDetectRule>>> listDetectRules() {
        return ResponseEntity.ok(Response.success(aiSensitiveWordService.listDetectRules()));
    }

    @PostMapping("/detect-rules")
    @Operation(summary = "Create or update an AI sensitive detect rule")
    public ResponseEntity<Response<AiSensitiveDetectRule>> saveDetectRule(@RequestBody AiSensitiveDetectRule rule) {
        return ResponseEntity.ok(Response.success(aiSensitiveWordService.saveDetectRule(rule)));
    }

    @PutMapping("/detect-rules/{id}")
    @Operation(summary = "Update an AI sensitive detect rule")
    public ResponseEntity<Response<AiSensitiveDetectRule>> updateDetectRule(@PathVariable("id") Long id,
        @RequestBody AiSensitiveDetectRule rule) {
        rule.setId(id);
        return ResponseEntity.ok(Response.success(aiSensitiveWordService.saveDetectRule(rule)));
    }

    @DeleteMapping("/detect-rules/{id}")
    @Operation(summary = "Delete an AI sensitive detect rule")
    public ResponseEntity<Response<Void>> deleteDetectRule(@PathVariable("id") Long id) {
        aiSensitiveWordService.deleteDetectRule(id);
        return ResponseEntity.noContent().build();
    }

    @GetMapping("/replace-rules")
    @Operation(summary = "List AI sensitive replace rules")
    public ResponseEntity<Response<List<AiSensitiveReplaceRule>>> listReplaceRules() {
        return ResponseEntity.ok(Response.success(aiSensitiveWordService.listReplaceRules()));
    }

    @PostMapping("/replace-rules")
    @Operation(summary = "Create or update an AI sensitive replace rule")
    public ResponseEntity<Response<AiSensitiveReplaceRule>> saveReplaceRule(@RequestBody AiSensitiveReplaceRule rule) {
        return ResponseEntity.ok(Response.success(aiSensitiveWordService.saveReplaceRule(rule)));
    }

    @PutMapping("/replace-rules/{id}")
    @Operation(summary = "Update an AI sensitive replace rule")
    public ResponseEntity<Response<AiSensitiveReplaceRule>> updateReplaceRule(@PathVariable("id") Long id,
        @RequestBody AiSensitiveReplaceRule rule) {
        rule.setId(id);
        return ResponseEntity.ok(Response.success(aiSensitiveWordService.saveReplaceRule(rule)));
    }

    @DeleteMapping("/replace-rules/{id}")
    @Operation(summary = "Delete an AI sensitive replace rule")
    public ResponseEntity<Response<Void>> deleteReplaceRule(@PathVariable("id") Long id) {
        aiSensitiveWordService.deleteReplaceRule(id);
        return ResponseEntity.noContent().build();
    }

    @GetMapping("/audits")
    @Operation(summary = "List AI sensitive block audits")
    public ResponseEntity<Response<List<AiSensitiveBlockAudit>>> listAudits(
        @RequestParam(required = false) String consumerName,
        @RequestParam(required = false) String displayName,
        @RequestParam(required = false) String routeName,
        @RequestParam(required = false) String matchType,
        @RequestParam(required = false) String startTime,
        @RequestParam(required = false) String endTime,
        @RequestParam(required = false) Integer limit) {
        return ResponseEntity.ok(Response.success(
            aiSensitiveWordService.listAudits(consumerName, displayName, routeName, matchType, startTime, endTime,
                limit)));
    }

    @GetMapping("/system-config")
    @Operation(summary = "Get AI sensitive system dictionary config")
    public ResponseEntity<Response<AiSensitiveSystemConfig>> getSystemConfig() {
        return ResponseEntity.ok(Response.success(aiSensitiveWordService.getSystemConfig()));
    }

    @PutMapping("/system-config")
    @Operation(summary = "Update AI sensitive system dictionary config")
    public ResponseEntity<Response<AiSensitiveSystemConfig>> updateSystemConfig(
        @RequestBody AiSensitiveSystemConfig config) {
        return ResponseEntity.ok(Response.success(aiSensitiveWordService.saveSystemConfig(config)));
    }

    @GetMapping("/status")
    @Operation(summary = "Get AI sensitive word status")
    public ResponseEntity<Response<AiSensitiveStatus>> getStatus() {
        return ResponseEntity.ok(Response.success(aiSensitiveWordService.getStatus()));
    }

    @PostMapping("/reconcile")
    @Operation(summary = "Trigger AI sensitive word reconcile")
    public ResponseEntity<Response<AiSensitiveStatus>> reconcile() {
        return ResponseEntity.ok(Response.success(aiSensitiveWordService.reconcile()));
    }
}
