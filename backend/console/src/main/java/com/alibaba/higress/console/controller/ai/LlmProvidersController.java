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
package com.alibaba.higress.console.controller.ai;

import javax.annotation.Resource;
import javax.validation.constraints.NotBlank;

import java.util.List;
import java.util.Map;

import org.apache.commons.lang3.StringUtils;
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

import com.alibaba.higress.console.controller.dto.PaginatedResponse;
import com.alibaba.higress.console.controller.dto.Response;
import com.alibaba.higress.console.controller.util.ControllerUtil;
import com.alibaba.higress.console.service.portal.PortalModelPricingJdbcService;
import com.alibaba.higress.sdk.exception.ValidationException;
import com.alibaba.higress.sdk.model.CommonPageQuery;
import com.alibaba.higress.sdk.model.PaginatedResult;
import com.alibaba.higress.sdk.model.ai.LlmProvider;
import com.alibaba.higress.sdk.service.ai.LlmProviderService;

import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.responses.ApiResponse;
import io.swagger.v3.oas.annotations.responses.ApiResponses;
import io.swagger.v3.oas.annotations.tags.Tag;
import lombok.extern.slf4j.Slf4j;

@Slf4j
@RestController("LlmProvidersController")
@RequestMapping("/v1/ai/providers")
@Validated
@Tag(name = "LLM Provider APIs")
public class LlmProvidersController {

    private static final String PORTAL_MODEL_META_KEY = "portalModelMeta";
    private static final String CURRENCY_CNY = "CNY";

    private LlmProviderService llmProviderService;
    private PortalModelPricingJdbcService portalModelPricingJdbcService;

    @Resource
    public void setLlmProviderService(LlmProviderService llmProviderService) {
        this.llmProviderService = llmProviderService;
    }

    @Resource
    public void setPortalModelPricingJdbcService(PortalModelPricingJdbcService portalModelPricingJdbcService) {
        this.portalModelPricingJdbcService = portalModelPricingJdbcService;
    }

    @GetMapping
    @Operation(summary = "List LLM providers")
    @ApiResponses(value = {@ApiResponse(responseCode = "200", description = "Providers listed successfully"),
        @ApiResponse(responseCode = "500", description = "Internal server error")})
    public ResponseEntity<PaginatedResponse<LlmProvider>> list(@RequestParam(required = false) CommonPageQuery query) {
        PaginatedResult<LlmProvider> providers = llmProviderService.list(query);
        return ControllerUtil.buildResponseEntity(providers);
    }

    @PostMapping
    @Operation(summary = "Add a new LLM provider")
    @ApiResponses(value = {@ApiResponse(responseCode = "200", description = "Route added successfully"),
        @ApiResponse(responseCode = "400", description = "Route data is not valid"),
        @ApiResponse(responseCode = "409", description = "Route already existed with the same name."),
        @ApiResponse(responseCode = "500", description = "Internal server error")})
    public ResponseEntity<Response<LlmProvider>> add(@RequestBody LlmProvider provider) {
        provider.validate(false);
        validatePortalModelMeta(provider);
        LlmProvider newProvider = addOrUpdateWithPortalSync(provider);
        return ControllerUtil.buildResponseEntity(newProvider);
    }

    @GetMapping(value = "/{name}")
    @Operation(summary = "Get LLM provider by name")
    @ApiResponses(value = {@ApiResponse(responseCode = "200", description = "Provider found"),
        @ApiResponse(responseCode = "404", description = "Provider not found"),
        @ApiResponse(responseCode = "500", description = "Internal server error")})
    public ResponseEntity<Response<LlmProvider>> query(@PathVariable("name") @NotBlank String name) {
        LlmProvider provider = llmProviderService.query(name);
        return ControllerUtil.buildResponseEntity(provider);
    }

    @PutMapping("/{name}")
    @Operation(summary = "Update an existed provider")
    @ApiResponses(value = {@ApiResponse(responseCode = "200", description = "Provider updated successfully"),
        @ApiResponse(responseCode = "400",
            description = "Provider data is not valid or provider name in the URL doesn't match the one in the body."),
        @ApiResponse(responseCode = "500", description = "Internal server error")})
    public ResponseEntity<Response<LlmProvider>> put(@PathVariable("name") @NotBlank String name,
        @RequestBody LlmProvider provider) {
        if (StringUtils.isNotEmpty(provider.getName())) {
            provider.setName(name);
        } else if (!StringUtils.equals(name, provider.getName())) {
            throw new ValidationException("Provider name in the URL doesn't match the one in the body.");
        }
        provider.validate(false);
        validatePortalModelMeta(provider);
        LlmProvider updatedProvider = addOrUpdateWithPortalSync(provider);
        return ControllerUtil.buildResponseEntity(updatedProvider);
    }

    @DeleteMapping("/{name}")
    @Operation(summary = "Delete an LLM provider")
    @ApiResponses(value = {@ApiResponse(responseCode = "204", description = "Provider deleted successfully"),
        @ApiResponse(responseCode = "500", description = "Internal server error")})
    public ResponseEntity<Response<LlmProvider>> delete(@PathVariable("name") @NotBlank String name) {
        deleteWithPortalSync(name);
        return ResponseEntity.noContent().build();
    }

    private void validatePortalModelMeta(LlmProvider provider) {
        if (provider == null || provider.getRawConfigs() == null || provider.getRawConfigs().isEmpty()) {
            throw new ValidationException("rawConfigs.portalModelMeta is required.");
        }
        Object metaObj = provider.getRawConfigs().get(PORTAL_MODEL_META_KEY);
        if (metaObj == null) {
            throw new ValidationException("rawConfigs.portalModelMeta is required.");
        }
        if (!(metaObj instanceof Map)) {
            throw new ValidationException("rawConfigs.portalModelMeta must be an object.");
        }
        @SuppressWarnings("unchecked")
        Map<String, Object> portalModelMeta = (Map<String, Object>)metaObj;
        validateStringField(portalModelMeta, "intro");
        validateStringListField(portalModelMeta, "tags");

        Object capabilitiesObj = portalModelMeta.get("capabilities");
        if (capabilitiesObj != null) {
            Map<String, Object> capabilities = requireMap(capabilitiesObj, "rawConfigs.portalModelMeta.capabilities");
            validateStringListField(capabilities, "modalities", "rawConfigs.portalModelMeta.capabilities");
            validateStringListField(capabilities, "features", "rawConfigs.portalModelMeta.capabilities");
        }

        Object pricingObj = portalModelMeta.get("pricing");
        if (pricingObj == null) {
            throw new ValidationException("rawConfigs.portalModelMeta.pricing is required.");
        }
        Map<String, Object> pricing = requireMap(pricingObj, "rawConfigs.portalModelMeta.pricing");
        validateStringField(pricing, "currency", "rawConfigs.portalModelMeta.pricing");
        String currency = StringUtils.trimToEmpty((String)pricing.get("currency"));
        if (StringUtils.isNotBlank(currency) && !StringUtils.equalsIgnoreCase(currency, CURRENCY_CNY)) {
            throw new ValidationException("rawConfigs.portalModelMeta.pricing.currency must be CNY.");
        }
        requireNumberField(pricing, "inputPer1K", "rawConfigs.portalModelMeta.pricing", false);
        requireNumberField(pricing, "outputPer1K", "rawConfigs.portalModelMeta.pricing", false);
        pricing.put("currency", CURRENCY_CNY);

        Object limitsObj = portalModelMeta.get("limits");
        if (limitsObj != null) {
            Map<String, Object> limits = requireMap(limitsObj, "rawConfigs.portalModelMeta.limits");
            validateNumberField(limits, "rpm", "rawConfigs.portalModelMeta.limits", true);
            validateNumberField(limits, "tpm", "rawConfigs.portalModelMeta.limits", true);
            validateNumberField(limits, "contextWindow", "rawConfigs.portalModelMeta.limits", true);
        }
    }

    private Map<String, Object> requireMap(Object value, String path) {
        if (!(value instanceof Map)) {
            throw new ValidationException(path + " must be an object.");
        }
        @SuppressWarnings("unchecked")
        Map<String, Object> result = (Map<String, Object>)value;
        return result;
    }

    private void validateStringField(Map<String, Object> container, String fieldName) {
        validateStringField(container, fieldName, "rawConfigs.portalModelMeta");
    }

    private void validateStringField(Map<String, Object> container, String fieldName, String parentPath) {
        Object value = container.get(fieldName);
        if (value == null) {
            return;
        }
        if (!(value instanceof String)) {
            throw new ValidationException(parentPath + "." + fieldName + " must be a string.");
        }
    }

    private void validateStringListField(Map<String, Object> container, String fieldName) {
        validateStringListField(container, fieldName, "rawConfigs.portalModelMeta");
    }

    private void validateStringListField(Map<String, Object> container, String fieldName, String parentPath) {
        Object value = container.get(fieldName);
        if (value == null) {
            return;
        }
        if (!(value instanceof List)) {
            throw new ValidationException(parentPath + "." + fieldName + " must be an array.");
        }
        @SuppressWarnings("unchecked")
        List<Object> values = (List<Object>)value;
        for (Object item : values) {
            if (!(item instanceof String)) {
                throw new ValidationException(parentPath + "." + fieldName + " items must be strings.");
            }
        }
    }

    private void validateNumberField(Map<String, Object> container, String fieldName, String parentPath,
        boolean requireInteger) {
        Object value = container.get(fieldName);
        if (value == null) {
            return;
        }
        Double numberValue = parseNumber(value);
        if (numberValue == null) {
            throw new ValidationException(parentPath + "." + fieldName + " must be a number.");
        }
        if (numberValue < 0) {
            throw new ValidationException(parentPath + "." + fieldName + " cannot be negative.");
        }
        if (requireInteger && numberValue % 1 != 0) {
            throw new ValidationException(parentPath + "." + fieldName + " must be an integer.");
        }
    }

    private void requireNumberField(Map<String, Object> container, String fieldName, String parentPath,
        boolean requireInteger) {
        Object value = container.get(fieldName);
        if (value == null) {
            throw new ValidationException(parentPath + "." + fieldName + " is required.");
        }
        validateNumberField(container, fieldName, parentPath, requireInteger);
    }

    private Double parseNumber(Object value) {
        if (value instanceof Number) {
            return ((Number)value).doubleValue();
        }
        if (value instanceof String) {
            String text = StringUtils.trimToNull((String)value);
            if (text == null) {
                return null;
            }
            try {
                return Double.parseDouble(text);
            } catch (NumberFormatException ex) {
                return null;
            }
        }
        return null;
    }

    private LlmProvider addOrUpdateWithPortalSync(LlmProvider provider) {
        LlmProvider existedProvider = llmProviderService.query(provider.getName());
        LlmProvider savedProvider = llmProviderService.addOrUpdate(provider);
        try {
            portalModelPricingJdbcService.upsertProvider(savedProvider);
            return savedProvider;
        } catch (RuntimeException ex) {
            rollbackProviderMutation(savedProvider.getName(), existedProvider);
            throw ex;
        }
    }

    private void deleteWithPortalSync(String providerName) {
        LlmProvider existedProvider = llmProviderService.query(providerName);
        llmProviderService.delete(providerName);
        try {
            portalModelPricingJdbcService.disableProvider(providerName);
        } catch (RuntimeException ex) {
            rollbackProviderMutation(providerName, existedProvider);
            throw ex;
        }
    }

    private void rollbackProviderMutation(String providerName, LlmProvider existedProvider) {
        try {
            if (existedProvider == null) {
                llmProviderService.delete(providerName);
            } else {
                llmProviderService.addOrUpdate(existedProvider);
            }
        } catch (RuntimeException rollbackEx) {
            log.error("Failed to rollback provider mutation for {} after Portal pricing sync failed.", providerName,
                rollbackEx);
        }
    }
}
