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
package com.alibaba.higress.sdk.model;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collections;
import java.util.LinkedHashSet;
import java.util.List;
import java.util.Locale;

import org.apache.commons.lang3.StringUtils;

import com.alibaba.higress.sdk.exception.ValidationException;

import io.swagger.v3.oas.annotations.media.Schema;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
@Schema(description = "Route auth configuration")
public class RouteAuthConfig {

    public static final String ANNOTATION_ALLOWED_CONSUMER_LEVELS = "higress.io/auth-consumer-levels";

    public static final String USER_LEVEL_NORMAL = "normal";
    public static final String USER_LEVEL_PLUS = "plus";
    public static final String USER_LEVEL_PRO = "pro";
    public static final String USER_LEVEL_ULTRA = "ultra";

    private static final List<String> SUPPORTED_USER_LEVELS = Collections.unmodifiableList(
        Arrays.asList(USER_LEVEL_NORMAL, USER_LEVEL_PLUS, USER_LEVEL_PRO, USER_LEVEL_ULTRA));

    @Schema(description = "Whether auth is enabled")
    private Boolean enabled;
    @Schema(description = "Allowed credential types")
    private List<String> allowedCredentialTypes;
    @Schema(description = "Allowed consumer names")
    private List<String> allowedConsumers;
    @Schema(description = "Allowed consumer levels")
    private List<String> allowedConsumerLevels;

    public void validate() {
        allowedConsumerLevels = normalizeAllowedConsumerLevels(allowedConsumerLevels);
    }

    public static List<String> normalizeAllowedConsumerLevels(List<String> levels) {
        if (levels == null || levels.isEmpty()) {
            return Collections.emptyList();
        }
        LinkedHashSet<String> normalized = new LinkedHashSet<>();
        for (String level : levels) {
            String value = StringUtils.trimToEmpty(level).toLowerCase(Locale.ROOT);
            if (StringUtils.isBlank(value)) {
                continue;
            }
            if (!SUPPORTED_USER_LEVELS.contains(value)) {
                throw new ValidationException("allowedConsumerLevels must be one of normal/plus/pro/ultra.");
            }
            normalized.add(value);
        }
        return new ArrayList<>(normalized);
    }

    public static List<String> parseAllowedConsumerLevels(String rawValue) {
        if (StringUtils.isBlank(rawValue)) {
            return Collections.emptyList();
        }
        String[] split = rawValue.split(",");
        List<String> values = new ArrayList<>(split.length);
        for (String value : split) {
            if (StringUtils.isNotBlank(value)) {
                values.add(value);
            }
        }
        return normalizeAllowedConsumerLevels(values);
    }

    public static String encodeAllowedConsumerLevels(List<String> levels) {
        List<String> normalized = normalizeAllowedConsumerLevels(levels);
        if (normalized.isEmpty()) {
            return StringUtils.EMPTY;
        }
        return String.join(",", normalized);
    }
}
