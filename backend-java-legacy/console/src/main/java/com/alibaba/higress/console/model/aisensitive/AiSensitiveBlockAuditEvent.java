package com.alibaba.higress.console.model.aisensitive;

import java.time.LocalDateTime;

import io.swagger.v3.oas.annotations.media.Schema;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
@Schema(description = "AI sensitive block audit event")
public class AiSensitiveBlockAuditEvent {

    private String requestId;
    private String routeName;
    private String consumerName;
    private LocalDateTime blockedAt;
    private String blockedBy;
    private String requestPhase;
    private String blockedReasonJson;
    private String matchType;
    private String matchedRule;
    private String matchedExcerpt;
}
