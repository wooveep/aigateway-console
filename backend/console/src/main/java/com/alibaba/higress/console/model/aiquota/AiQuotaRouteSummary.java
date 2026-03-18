package com.alibaba.higress.console.model.aiquota;

import java.util.List;

import io.swagger.v3.oas.annotations.media.Schema;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
@Schema(description = "AI route summary for quota management")
public class AiQuotaRouteSummary {

    @Schema(description = "AI route name")
    private String routeName;

    @Schema(description = "Bound domains")
    private List<String> domains;

    @Schema(description = "Path match value")
    private String path;

    @Schema(description = "Redis key prefix")
    private String redisKeyPrefix;

    @Schema(description = "Admin consumer configured on ai-quota")
    private String adminConsumer;

    @Schema(description = "Admin path configured on ai-quota")
    private String adminPath;

    @Schema(description = "Configured schedule rule count")
    private int scheduleRuleCount;
}
