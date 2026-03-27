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
@Schema(description = "Consumer quota snapshot")
public class AiQuotaConsumerQuota {

    @Schema(description = "Consumer name")
    private String consumerName;

    @Schema(description = "Current quota")
    private long quota;
}
