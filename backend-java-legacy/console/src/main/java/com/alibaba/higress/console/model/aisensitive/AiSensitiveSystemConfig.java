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
@Schema(description = "AI sensitive system dictionary config")
public class AiSensitiveSystemConfig {

    private Boolean systemDenyEnabled;
    private String dictionaryText;
    private String updatedAt;
    private String updatedBy;
}
