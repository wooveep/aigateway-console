package com.alibaba.higress.console.controller;

import java.util.List;

import javax.annotation.Resource;
import javax.validation.constraints.NotBlank;

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
import com.alibaba.higress.console.model.aiquota.AiQuotaConsumerQuota;
import com.alibaba.higress.console.model.aiquota.AiQuotaMenuState;
import com.alibaba.higress.console.model.aiquota.AiQuotaRouteSummary;
import com.alibaba.higress.console.model.aiquota.AiQuotaScheduleRule;
import com.alibaba.higress.console.model.aiquota.AiQuotaScheduleRuleRequest;
import com.alibaba.higress.console.model.aiquota.AiQuotaUserPolicy;
import com.alibaba.higress.console.model.aiquota.AiQuotaUserPolicyRequest;
import com.alibaba.higress.console.model.aiquota.AiQuotaValueRequest;
import com.alibaba.higress.console.service.AiQuotaService;

import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.responses.ApiResponse;
import io.swagger.v3.oas.annotations.responses.ApiResponses;
import io.swagger.v3.oas.annotations.tags.Tag;

@RestController("AiQuotaController")
@RequestMapping("/v1/ai/quotas")
@Validated
@Tag(name = "AI Quota APIs")
public class AiQuotaController {

    private AiQuotaService aiQuotaService;

    @Resource
    public void setAiQuotaService(AiQuotaService aiQuotaService) {
        this.aiQuotaService = aiQuotaService;
    }

    @GetMapping("/menu-state")
    @Operation(summary = "Get ai-quota menu state")
    @ApiResponses(value = {@ApiResponse(responseCode = "200", description = "Menu state retrieved successfully")})
    public ResponseEntity<Response<AiQuotaMenuState>> getMenuState() {
        return ResponseEntity.ok(Response.success(aiQuotaService.getMenuState()));
    }

    @GetMapping("/routes")
    @Operation(summary = "List AI routes with ai-quota enabled")
    @ApiResponses(value = {@ApiResponse(responseCode = "200", description = "Routes listed successfully")})
    public ResponseEntity<Response<List<AiQuotaRouteSummary>>> listRoutes() {
        return ResponseEntity.ok(Response.success(aiQuotaService.listEnabledRoutes()));
    }

    @GetMapping("/routes/{routeName}/consumers")
    @Operation(summary = "List consumer quotas for an AI route")
    @ApiResponses(value = {@ApiResponse(responseCode = "200", description = "Consumer quotas listed successfully")})
    public ResponseEntity<Response<List<AiQuotaConsumerQuota>>> listConsumerQuotas(
        @PathVariable("routeName") @NotBlank String routeName) {
        return ResponseEntity.ok(Response.success(aiQuotaService.listConsumerQuotas(routeName)));
    }

    @PutMapping("/routes/{routeName}/consumers/{consumerName}/quota")
    @Operation(summary = "Refresh quota for a consumer")
    @ApiResponses(value = {@ApiResponse(responseCode = "200", description = "Quota refreshed successfully")})
    public ResponseEntity<Response<AiQuotaConsumerQuota>> refreshQuota(
        @PathVariable("routeName") @NotBlank String routeName,
        @PathVariable("consumerName") @NotBlank String consumerName,
        @RequestBody AiQuotaValueRequest request) {
        if (request == null || request.getValue() == null) {
            throw new IllegalArgumentException("quota value cannot be null.");
        }
        return ResponseEntity.ok(Response.success(aiQuotaService.refreshQuota(routeName, consumerName, request.getValue())));
    }

    @PostMapping("/routes/{routeName}/consumers/{consumerName}/delta")
    @Operation(summary = "Increase or decrease quota for a consumer")
    @ApiResponses(value = {@ApiResponse(responseCode = "200", description = "Quota adjusted successfully")})
    public ResponseEntity<Response<AiQuotaConsumerQuota>> deltaQuota(
        @PathVariable("routeName") @NotBlank String routeName,
        @PathVariable("consumerName") @NotBlank String consumerName,
        @RequestBody AiQuotaValueRequest request) {
        if (request == null || request.getValue() == null) {
            throw new IllegalArgumentException("delta value cannot be null.");
        }
        return ResponseEntity.ok(Response.success(aiQuotaService.deltaQuota(routeName, consumerName, request.getValue())));
    }

    @GetMapping("/routes/{routeName}/consumers/{consumerName}/policy")
    @Operation(summary = "Get user-level quota policy for a consumer")
    @ApiResponses(value = {@ApiResponse(responseCode = "200", description = "User quota policy retrieved successfully")})
    public ResponseEntity<Response<AiQuotaUserPolicy>> getUserPolicy(
        @PathVariable("routeName") @NotBlank String routeName,
        @PathVariable("consumerName") @NotBlank String consumerName) {
        return ResponseEntity.ok(Response.success(aiQuotaService.getUserPolicy(routeName, consumerName)));
    }

    @PutMapping("/routes/{routeName}/consumers/{consumerName}/policy")
    @Operation(summary = "Create or update user-level quota policy for a consumer")
    @ApiResponses(value = {@ApiResponse(responseCode = "200", description = "User quota policy saved successfully")})
    public ResponseEntity<Response<AiQuotaUserPolicy>> saveUserPolicy(
        @PathVariable("routeName") @NotBlank String routeName,
        @PathVariable("consumerName") @NotBlank String consumerName,
        @RequestBody AiQuotaUserPolicyRequest request) {
        if (request == null) {
            throw new IllegalArgumentException("user quota policy request cannot be null.");
        }
        return ResponseEntity.ok(Response.success(aiQuotaService.saveUserPolicy(routeName, consumerName, request)));
    }

    @GetMapping("/routes/{routeName}/schedules")
    @Operation(summary = "List scheduled quota rules")
    @ApiResponses(value = {@ApiResponse(responseCode = "200", description = "Schedule rules listed successfully")})
    public ResponseEntity<Response<List<AiQuotaScheduleRule>>> listScheduleRules(
        @PathVariable("routeName") @NotBlank String routeName,
        @RequestParam(required = false) @Parameter(description = "Filter by consumer name") String consumerName) {
        return ResponseEntity.ok(Response.success(aiQuotaService.listScheduleRules(routeName, consumerName)));
    }

    @PutMapping("/routes/{routeName}/schedules")
    @Operation(summary = "Create or update a scheduled quota rule")
    @ApiResponses(value = {@ApiResponse(responseCode = "200", description = "Schedule rule saved successfully")})
    public ResponseEntity<Response<AiQuotaScheduleRule>> saveScheduleRule(
        @PathVariable("routeName") @NotBlank String routeName,
        @RequestBody AiQuotaScheduleRuleRequest request) {
        return ResponseEntity.ok(Response.success(aiQuotaService.saveScheduleRule(routeName, request)));
    }

    @DeleteMapping("/routes/{routeName}/schedules/{ruleId}")
    @Operation(summary = "Delete a scheduled quota rule")
    @ApiResponses(value = {@ApiResponse(responseCode = "204", description = "Schedule rule deleted successfully")})
    public ResponseEntity<Response<Void>> deleteScheduleRule(
        @PathVariable("routeName") @NotBlank String routeName,
        @PathVariable("ruleId") @NotBlank String ruleId) {
        aiQuotaService.deleteScheduleRule(routeName, ruleId);
        return ResponseEntity.noContent().build();
    }
}
