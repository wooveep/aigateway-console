/*
 * Copyright (c) 2022-2023 Alibaba Group Holding Ltd.
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
package com.alibaba.higress.console.controller;

import java.util.Arrays;
import java.util.Collections;
import java.util.List;
import java.util.Map;
import java.util.UUID;
import java.util.stream.Collectors;

import javax.annotation.Resource;
import javax.validation.ValidationException;
import javax.validation.constraints.NotBlank;

import org.apache.commons.collections4.CollectionUtils;
import org.apache.commons.lang3.StringUtils;
import org.springdoc.api.annotations.ParameterObject;
import org.springframework.http.ResponseEntity;
import org.springframework.validation.annotation.Validated;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PatchMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import com.alibaba.higress.console.controller.dto.PaginatedResponse;
import com.alibaba.higress.console.controller.dto.Response;
import com.alibaba.higress.console.controller.util.ControllerUtil;
import com.alibaba.higress.console.model.portal.PortalUserRecord;
import com.alibaba.higress.console.service.portal.PortalUserJdbcService;
import com.alibaba.higress.sdk.model.CommonPageQuery;
import com.alibaba.higress.sdk.model.PaginatedResult;
import com.alibaba.higress.sdk.model.consumer.Consumer;
import com.alibaba.higress.sdk.model.consumer.CredentialType;
import com.alibaba.higress.sdk.model.consumer.KeyAuthCredential;
import com.alibaba.higress.sdk.model.consumer.KeyAuthCredentialSource;
import com.alibaba.higress.sdk.service.consumer.ConsumerService;

import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.responses.ApiResponse;
import io.swagger.v3.oas.annotations.responses.ApiResponses;
import io.swagger.v3.oas.annotations.tags.Tag;

@RestController("ConsumersController")
@RequestMapping("/v1/consumers")
@Validated
@Tag(name = "Consumer APIs")
public class ConsumersController {

    private static final String STATUS_ACTIVE = "active";
    private static final String STATUS_DISABLED = "disabled";
    private static final String STATUS_PENDING = "pending";

    private ConsumerService consumerService;
    private PortalUserJdbcService portalUserJdbcService;

    @Resource
    public void setConsumerService(ConsumerService consumerService) {
        this.consumerService = consumerService;
    }

    @Resource
    public void setPortalUserJdbcService(PortalUserJdbcService portalUserJdbcService) {
        this.portalUserJdbcService = portalUserJdbcService;
    }

    @GetMapping
    @Operation(summary = "List consumers")
    @ApiResponses(value = {@ApiResponse(responseCode = "200", description = "Consumers listed successfully"),
        @ApiResponse(responseCode = "500", description = "Internal server error")})
    public ResponseEntity<PaginatedResponse<Consumer>> list(@ParameterObject CommonPageQuery query) {
        PaginatedResult<Consumer> consumers = consumerService.list(query);
        enrichConsumers(consumers.getData());
        return ControllerUtil.buildResponseEntity(consumers);
    }

    @PostMapping
    @Operation(summary = "Add a consumer")
    @ApiResponses(value = {@ApiResponse(responseCode = "200", description = "Consumer added successfully"),
        @ApiResponse(responseCode = "400", description = "Consumer data is not valid"),
        @ApiResponse(responseCode = "500", description = "Internal server error")})
    public ResponseEntity<Response<Consumer>> add(@RequestBody Consumer consumer) {
        consumer.validate(false);
        Consumer newConsumer = consumerService.addOrUpdate(consumer);
        upsertPortalUser(newConsumer, consumer, true);
        enrichConsumers(Collections.singletonList(newConsumer));
        return ControllerUtil.buildResponseEntity(newConsumer);
    }

    @GetMapping(value = "/{name}")
    @Operation(summary = "Get consumer by name")
    @ApiResponses(value = {@ApiResponse(responseCode = "200", description = "Consumer found"),
        @ApiResponse(responseCode = "404", description = "Consumer not found"),
        @ApiResponse(responseCode = "500", description = "Internal server error")})
    public ResponseEntity<Response<Consumer>> query(@PathVariable("name") @NotBlank String name) {
        Consumer consumer = consumerService.query(name);
        enrichConsumers(Collections.singletonList(consumer));
        return ControllerUtil.buildResponseEntity(consumer);
    }

    @GetMapping("/departments")
    @Operation(summary = "List departments")
    @ApiResponses(value = {@ApiResponse(responseCode = "200", description = "Departments listed successfully"),
        @ApiResponse(responseCode = "500", description = "Internal server error")})
    public ResponseEntity<Response<java.util.List<String>>> listDepartments() {
        return ControllerUtil.buildResponseEntity(consumerService.listDepartments());
    }

    @PostMapping("/departments")
    @Operation(summary = "Add a department")
    @ApiResponses(value = {@ApiResponse(responseCode = "204", description = "Department added successfully"),
        @ApiResponse(responseCode = "400", description = "Department data is not valid"),
        @ApiResponse(responseCode = "500", description = "Internal server error")})
    public ResponseEntity<Void> addDepartment(@RequestBody DepartmentRequest request) {
        if (request == null || StringUtils.isBlank(request.getName())) {
            throw new ValidationException("Department name cannot be blank.");
        }
        consumerService.addDepartment(request.getName());
        return ResponseEntity.noContent().build();
    }

    @PutMapping("/{name}")
    @Operation(summary = "Update an existed consumer")
    @ApiResponses(value = {@ApiResponse(responseCode = "200", description = "Consumer updated successfully"),
        @ApiResponse(responseCode = "400", description = "Consumer data is not valid"),
        @ApiResponse(responseCode = "500", description = "Internal server error")})
    public ResponseEntity<Response<Consumer>> put(@PathVariable("name") @NotBlank String name,
        @RequestBody Consumer consumer) {
        if (StringUtils.isEmpty(consumer.getName())) {
            consumer.setName(name);
        } else if (!StringUtils.equals(name, consumer.getName())) {
            throw new ValidationException("Consumer name in the URL doesn't match the one in the body.");
        }
        consumer.validate(true);
        Consumer updatedConsumer = consumerService.addOrUpdate(consumer);
        upsertPortalUser(updatedConsumer, consumer, false);
        enrichConsumers(Collections.singletonList(updatedConsumer));
        return ControllerUtil.buildResponseEntity(updatedConsumer);
    }

    @PatchMapping("/{name}/status")
    @Operation(summary = "Update portal user status")
    @ApiResponses(value = {@ApiResponse(responseCode = "200", description = "Consumer status updated successfully"),
        @ApiResponse(responseCode = "400", description = "Invalid status value"),
        @ApiResponse(responseCode = "500", description = "Internal server error")})
    public ResponseEntity<Response<Consumer>> updateStatus(@PathVariable("name") @NotBlank String name,
        @RequestBody ConsumerStatusRequest request) {
        if (request == null || StringUtils.isBlank(request.getStatus())) {
            throw new ValidationException("status cannot be blank.");
        }
        String status = request.getStatus().trim().toLowerCase();
        if (!Arrays.asList(STATUS_ACTIVE, STATUS_DISABLED, STATUS_PENDING).contains(status)) {
            throw new ValidationException("status must be active/disabled/pending.");
        }

        Consumer consumer = consumerService.query(name);
        if (consumer == null) {
            throw new ValidationException("Consumer not found: " + name);
        }

        if (STATUS_DISABLED.equals(status)) {
            revokeConsumerKeys(consumer);
            consumerService.addOrUpdate(consumer);
            portalUserJdbcService.disableAllApiKeys(name);
        }
        consumer.setPortalStatus(status);
        portalUserJdbcService.upsertFromConsumer(consumer, "console");

        Consumer updated = consumerService.query(name);
        enrichConsumers(Collections.singletonList(updated));
        return ControllerUtil.buildResponseEntity(updated);
    }

    @DeleteMapping("/{name}")
    @Operation(summary = "Delete a consumer")
    @ApiResponses(value = {@ApiResponse(responseCode = "204", description = "Consumer deleted successfully"),
        @ApiResponse(responseCode = "500", description = "Internal server error")})
    public ResponseEntity<Response<Consumer>> delete(@PathVariable("name") @NotBlank String name) {
        consumerService.delete(name);
        portalUserJdbcService.updateStatus(name, STATUS_DISABLED);
        portalUserJdbcService.disableAllApiKeys(name);
        return ResponseEntity.noContent().build();
    }

    private void revokeConsumerKeys(Consumer consumer) {
        if (consumer == null) {
            return;
        }
        String revokedValue = "revoked_" + UUID.randomUUID().toString().replace("-", "");
        if (CollectionUtils.isNotEmpty(consumer.getCredentials())) {
            for (Object credential : consumer.getCredentials()) {
                if (!(credential instanceof KeyAuthCredential)) {
                    continue;
                }
                KeyAuthCredential keyAuthCredential = (KeyAuthCredential) credential;
                keyAuthCredential.setValues(Collections.singletonList(revokedValue));
                if (StringUtils.isBlank(keyAuthCredential.getSource())) {
                    keyAuthCredential.setSource(KeyAuthCredentialSource.BEARER.name());
                }
                return;
            }
        }

        KeyAuthCredential keyAuthCredential = new KeyAuthCredential();
        keyAuthCredential.setType(CredentialType.KEY_AUTH);
        keyAuthCredential.setSource(KeyAuthCredentialSource.BEARER.name());
        keyAuthCredential.setValues(Collections.singletonList(revokedValue));
        consumer.setCredentials(Collections.singletonList(keyAuthCredential));
    }

    private void upsertPortalUser(Consumer target, Consumer reqConsumer, boolean forCreate) {
        if (target == null) {
            return;
        }
        if (reqConsumer != null) {
            target.setPortalDisplayName(reqConsumer.getPortalDisplayName());
            target.setPortalEmail(reqConsumer.getPortalEmail());
            target.setPortalStatus(reqConsumer.getPortalStatus());
            target.setPortalUserSource(reqConsumer.getPortalUserSource());
            target.setPortalPassword(reqConsumer.getPortalPassword());
        }
        PortalUserRecord record = portalUserJdbcService.upsertFromConsumer(target, "console");
        if (record != null) {
            applyPortalData(target, record);
            if (forCreate && StringUtils.isNotBlank(record.getTempPassword())) {
                target.setPortalTempPassword(record.getTempPassword());
            }
        }
        target.setPortalPassword(null);
    }

    private void enrichConsumers(List<Consumer> consumers) {
        if (CollectionUtils.isEmpty(consumers)) {
            return;
        }
        List<String> names = consumers.stream().filter(c -> c != null && StringUtils.isNotBlank(c.getName()))
            .map(Consumer::getName).collect(Collectors.toList());
        Map<String, PortalUserRecord> portalUsers = portalUserJdbcService.listByConsumerNames(names);
        for (Consumer consumer : consumers) {
            if (consumer == null || StringUtils.isBlank(consumer.getName())) {
                continue;
            }
            PortalUserRecord record = portalUsers.get(consumer.getName());
            if (record == null) {
                if (StringUtils.isBlank(consumer.getPortalStatus())) {
                    consumer.setPortalStatus(STATUS_PENDING);
                }
                continue;
            }
            applyPortalData(consumer, record);
        }
    }

    private static void applyPortalData(Consumer consumer, PortalUserRecord record) {
        consumer.setPortalStatus(record.getStatus());
        consumer.setPortalDisplayName(record.getDisplayName());
        consumer.setPortalEmail(record.getEmail());
        consumer.setPortalUserSource(record.getSource());
    }

    public static class DepartmentRequest {

        private String name;

        public String getName() {
            return name;
        }

        public void setName(String name) {
            this.name = name;
        }
    }

    public static class ConsumerStatusRequest {

        private String status;

        public String getStatus() {
            return status;
        }

        public void setStatus(String status) {
            this.status = status;
        }
    }
}
