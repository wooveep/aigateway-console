package com.alibaba.higress.console.service;

import java.util.List;

import com.alibaba.higress.console.model.aisensitive.AiSensitiveBlockAudit;
import com.alibaba.higress.console.model.aisensitive.AiSensitiveBlockAuditEvent;
import com.alibaba.higress.console.model.aisensitive.AiSensitiveDetectRule;
import com.alibaba.higress.console.model.aisensitive.AiSensitiveMenuState;
import com.alibaba.higress.console.model.aisensitive.AiSensitiveReplaceRule;
import com.alibaba.higress.console.model.aisensitive.AiSensitiveStatus;
import com.alibaba.higress.console.model.aisensitive.AiSensitiveSystemConfig;

public interface AiSensitiveWordService {

    List<AiSensitiveDetectRule> listDetectRules();

    AiSensitiveDetectRule saveDetectRule(AiSensitiveDetectRule rule);

    void deleteDetectRule(Long id);

    List<AiSensitiveReplaceRule> listReplaceRules();

    AiSensitiveReplaceRule saveReplaceRule(AiSensitiveReplaceRule rule);

    void deleteReplaceRule(Long id);

    List<AiSensitiveBlockAudit> listAudits(String consumerName, String displayName, String routeName, String matchType,
        String startTime, String endTime, Integer limit);

    AiSensitiveSystemConfig getSystemConfig();

    AiSensitiveSystemConfig saveSystemConfig(AiSensitiveSystemConfig config);

    AiSensitiveStatus getStatus();

    AiSensitiveMenuState getMenuState();

    AiSensitiveStatus reconcile();

    AiSensitiveBlockAudit ingestBlockedEvent(AiSensitiveBlockAuditEvent event);
}
