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
@Schema(description = "AI quota menu visibility state")
public class AiQuotaMenuState {

    @Schema(description = "Whether the AI quota menu should be shown")
    private boolean enabled;

    @Schema(description = "Number of AI routes with ai-quota enabled")
    private int enabledRouteCount;
}
