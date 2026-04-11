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
public class PortalUserRecord {

    private String consumerName;
    private String displayName;
    private String email;
    private String departmentId;
    private String parentConsumerName;
    private String userLevel;
    private String status;
    private String source;
    private Boolean deleted;
    private LocalDateTime lastLoginAt;
    private String tempPassword;
}
