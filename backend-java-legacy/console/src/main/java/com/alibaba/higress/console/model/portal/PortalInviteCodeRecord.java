package com.alibaba.higress.console.model.portal;

import java.time.LocalDateTime;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class PortalInviteCodeRecord {

    private String inviteCode;
    private String status;
    private LocalDateTime expiresAt;
    private String usedByConsumer;
    private LocalDateTime usedAt;
    private LocalDateTime createdAt;
}
