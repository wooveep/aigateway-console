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
public class PortalPasswordResetResult {

    private String consumerName;
    private String tempPassword;
    private LocalDateTime updatedAt;
}
