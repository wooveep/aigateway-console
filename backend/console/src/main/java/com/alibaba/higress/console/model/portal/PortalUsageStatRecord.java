package com.alibaba.higress.console.model.portal;

import io.swagger.v3.oas.annotations.media.Schema;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
@Schema(description = "Portal usage stats grouped by consumer and model")
public class PortalUsageStatRecord {

    @Schema(description = "Consumer name")
    private String consumerName;

    @Schema(description = "Model name")
    private String modelName;

    @Schema(description = "Request count")
    private long requestCount;

    @Schema(description = "Input tokens")
    private long inputTokens;

    @Schema(description = "Output tokens")
    private long outputTokens;

    @Schema(description = "Total tokens")
    private long totalTokens;
}
