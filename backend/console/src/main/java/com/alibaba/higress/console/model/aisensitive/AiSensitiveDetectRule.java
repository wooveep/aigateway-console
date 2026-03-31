package com.alibaba.higress.console.model.aisensitive;

import io.swagger.v3.oas.annotations.media.Schema;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
@Schema(description = "AI sensitive detect rule")
public class AiSensitiveDetectRule {

    private Long id;
    private String pattern;
    private String matchType;
    private String description;
    private Integer priority;
    private Boolean enabled;
    private String createdAt;
    private String updatedAt;
}
