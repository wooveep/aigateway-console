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
@Schema(description = "AI quota user-level policy")
public class AiQuotaUserPolicy {

    @Schema(description = "Consumer name")
    private String consumerName;

    @Schema(description = "Total cost limit in micro_yuan")
    private long limitTotal;

    @Schema(description = "5-hour rolling cost limit in micro_yuan")
    private long limit5h;

    @Schema(description = "Daily cost limit in micro_yuan")
    private long limitDaily;

    @Schema(description = "Daily reset mode")
    private String dailyResetMode;

    @Schema(description = "Daily reset time, such as 00:00")
    private String dailyResetTime;

    @Schema(description = "Weekly cost limit in micro_yuan")
    private long limitWeekly;

    @Schema(description = "Monthly cost limit in micro_yuan")
    private long limitMonthly;

    @Schema(description = "Soft reset start time in RFC3339 format")
    private String costResetAt;
}
