package com.alibaba.higress.console.controller;

import javax.annotation.Resource;
import javax.validation.constraints.NotBlank;

import org.springdoc.api.annotations.ParameterObject;
import org.springframework.http.ResponseEntity;
import org.springframework.validation.annotation.Validated;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PatchMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import com.alibaba.higress.console.controller.dto.PaginatedResponse;
import com.alibaba.higress.console.controller.dto.Response;
import com.alibaba.higress.console.controller.util.ControllerUtil;
import com.alibaba.higress.console.model.portal.PortalInviteCodePageQuery;
import com.alibaba.higress.console.model.portal.PortalInviteCodeRecord;
import com.alibaba.higress.console.service.portal.PortalInviteCodeJdbcService;

import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.responses.ApiResponse;
import io.swagger.v3.oas.annotations.responses.ApiResponses;
import io.swagger.v3.oas.annotations.tags.Tag;

@RestController("PortalInviteCodeController")
@RequestMapping("/v1/portal/invite-codes")
@Validated
@Tag(name = "Portal Invite Code APIs")
public class PortalInviteCodeController {

    private PortalInviteCodeJdbcService inviteCodeJdbcService;

    @Resource
    public void setInviteCodeJdbcService(PortalInviteCodeJdbcService inviteCodeJdbcService) {
        this.inviteCodeJdbcService = inviteCodeJdbcService;
    }

    @PostMapping
    @Operation(summary = "Create portal invite code")
    @ApiResponses(value = {@ApiResponse(responseCode = "200", description = "Invite code created successfully"),
        @ApiResponse(responseCode = "400", description = "Invalid request"),
        @ApiResponse(responseCode = "500", description = "Internal server error")})
    public ResponseEntity<Response<PortalInviteCodeRecord>>
        createInviteCode(@RequestBody(required = false) CreateInviteCodeRequest req) {
        Integer expiresInDays = req == null ? null : req.getExpiresInDays();
        return ControllerUtil.buildResponseEntity(inviteCodeJdbcService.createInviteCode(expiresInDays));
    }

    @GetMapping
    @Operation(summary = "List portal invite codes")
    @ApiResponses(value = {@ApiResponse(responseCode = "200", description = "Invite codes listed successfully"),
        @ApiResponse(responseCode = "500", description = "Internal server error")})
    public ResponseEntity<PaginatedResponse<PortalInviteCodeRecord>> listInviteCodes(
        @ParameterObject PortalInviteCodePageQuery query) {
        return ControllerUtil.buildResponseEntity(inviteCodeJdbcService.list(query));
    }

    @PatchMapping("/{code}")
    @Operation(summary = "Update invite code status")
    @ApiResponses(value = {@ApiResponse(responseCode = "200", description = "Invite code updated successfully"),
        @ApiResponse(responseCode = "400", description = "Invalid request"),
        @ApiResponse(responseCode = "500", description = "Internal server error")})
    public ResponseEntity<Response<PortalInviteCodeRecord>>
        updateStatus(@PathVariable("code") @NotBlank String code, @RequestBody UpdateInviteCodeStatusRequest req) {
        if (req == null || req.getStatus() == null) {
            throw new com.alibaba.higress.sdk.exception.ValidationException("status cannot be blank.");
        }
        String status = req.getStatus().trim().toLowerCase();
        if (!"active".equals(status) && !"disabled".equals(status)) {
            throw new com.alibaba.higress.sdk.exception.ValidationException("status must be 'active' or 'disabled'.");
        }
        return ControllerUtil.buildResponseEntity(inviteCodeJdbcService.updateInviteCodeStatus(code, status));
    }

    @lombok.Data
    public static class CreateInviteCodeRequest {

        private Integer expiresInDays;
    }

    @lombok.Data
    public static class UpdateInviteCodeStatusRequest {

        private String status;
    }
}
