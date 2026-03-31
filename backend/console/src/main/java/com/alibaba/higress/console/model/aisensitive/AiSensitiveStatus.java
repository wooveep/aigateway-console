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
@Schema(description = "AI sensitive word status")
public class AiSensitiveStatus {

    private Boolean dbEnabled;
    private Integer detectRuleCount;
    private Integer replaceRuleCount;
    private Integer auditRecordCount;
    private Boolean systemDenyEnabled;
    private Integer systemDictionaryWordCount;
    private String systemDictionaryUpdatedAt;
    private Integer projectedInstanceCount;
    private String lastReconciledAt;
    private String lastMigratedAt;
    private String lastError;
}
