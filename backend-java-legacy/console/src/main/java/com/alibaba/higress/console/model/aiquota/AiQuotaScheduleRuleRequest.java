package com.alibaba.higress.console.model.aiquota;

import io.swagger.v3.oas.annotations.media.Schema;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
@Schema(description = "AI quota schedule rule request")
public class AiQuotaScheduleRuleRequest {

    @Schema(description = "Rule id. Leave empty to create a new rule")
    private String id;

    @Schema(description = "Consumer name")
    private String consumerName;

    @Schema(description = "Action type: REFRESH or DELTA")
    private String action;

    @Schema(description = "Cron expression in Spring format")
    private String cron;

    @Schema(description = "Quota value for REFRESH or delta value for DELTA")
    private Long value;

    @Schema(description = "Whether the rule is enabled")
    private Boolean enabled;
}
