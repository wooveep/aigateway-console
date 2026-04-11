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
@Schema(description = "AI data masking menu visibility state")
public class AiSensitiveMenuState {

    @Schema(description = "Whether the AI data masking menu should be shown")
    private boolean enabled;

    @Schema(description = "Number of AI routes with ai-data-masking enabled")
    private int enabledRouteCount;
}
