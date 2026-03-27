package com.alibaba.higress.console.model.aiquota;

import io.swagger.v3.oas.annotations.media.Schema;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
@Schema(description = "Scheduled AI quota operation")
public class AiQuotaScheduleRule {

    @Schema(description = "Rule id")
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

    @Schema(description = "Rule creation timestamp in epoch milliseconds")
    private Long createdAt;

    @Schema(description = "Rule update timestamp in epoch milliseconds")
    private Long updatedAt;

    @Schema(description = "Last successful execution timestamp in epoch milliseconds")
    private Long lastAppliedAt;

    @Schema(description = "Last execution error, if any")
    private String lastError;
}
