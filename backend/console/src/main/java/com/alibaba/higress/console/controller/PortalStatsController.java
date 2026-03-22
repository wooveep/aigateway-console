package com.alibaba.higress.console.controller;

import java.util.List;

import javax.annotation.Resource;

import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.validation.annotation.Validated;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import com.alibaba.higress.console.controller.dto.Response;
import com.alibaba.higress.console.controller.util.ControllerUtil;
import com.alibaba.higress.console.model.portal.PortalUsageStatRecord;
import com.alibaba.higress.console.service.portal.PortalUsageStatsService;

import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.responses.ApiResponse;
import io.swagger.v3.oas.annotations.responses.ApiResponses;
import io.swagger.v3.oas.annotations.tags.Tag;

@RestController("PortalStatsController")
@RequestMapping("/v1/portal/stats")
@Validated
@Tag(name = "Portal Stats APIs")
public class PortalStatsController {

    private PortalUsageStatsService portalUsageStatsService;

    @Resource
    public void setPortalUsageStatsService(PortalUsageStatsService portalUsageStatsService) {
        this.portalUsageStatsService = portalUsageStatsService;
    }

    @GetMapping("/usage")
    @Operation(summary = "List usage stats grouped by consumer and model")
    @ApiResponses(value = {@ApiResponse(responseCode = "200", description = "Usage stats listed successfully"),
        @ApiResponse(responseCode = "500", description = "Internal server error")})
    public ResponseEntity<Response<List<PortalUsageStatRecord>>> listUsage(
        @RequestParam(required = false) Long from, @RequestParam(required = false) Long to) {
        try {
            List<PortalUsageStatRecord> result = portalUsageStatsService.listUsage(from, to);
            return ControllerUtil.buildResponseEntity(result);
        } catch (IllegalStateException ex) {
            return ResponseEntity.status(HttpStatus.SERVICE_UNAVAILABLE).body(Response.failure(ex.getMessage()));
        }
    }
}
