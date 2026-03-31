package com.alibaba.higress.console.service;

import java.util.Arrays;
import java.util.List;

import javax.annotation.Resource;

import org.apache.commons.lang3.StringUtils;
import org.springframework.stereotype.Service;

import com.alibaba.higress.console.model.aisensitive.AiSensitiveBlockAudit;
import com.alibaba.higress.console.model.aisensitive.AiSensitiveBlockAuditEvent;
import com.alibaba.higress.console.model.aisensitive.AiSensitiveDetectRule;
import com.alibaba.higress.console.model.aisensitive.AiSensitiveReplaceRule;
import com.alibaba.higress.console.model.aisensitive.AiSensitiveStatus;
import com.alibaba.higress.console.model.aisensitive.AiSensitiveSystemConfig;
import com.alibaba.higress.console.model.User;
import com.alibaba.higress.console.model.portal.PortalUserRecord;
import com.alibaba.higress.console.service.portal.AiSensitiveWordJdbcService;
import com.alibaba.higress.console.service.portal.PortalUserJdbcService;
import com.alibaba.higress.sdk.exception.ValidationException;

@Service
public class AiSensitiveWordServiceImpl implements AiSensitiveWordService {

    private AiSensitiveWordJdbcService aiSensitiveWordJdbcService;
    private PortalUserJdbcService portalUserJdbcService;
    private AiSensitiveWordProjectionService projectionService;

    @Resource
    public void setAiSensitiveWordJdbcService(AiSensitiveWordJdbcService aiSensitiveWordJdbcService) {
        this.aiSensitiveWordJdbcService = aiSensitiveWordJdbcService;
    }

    @Resource
    public void setPortalUserJdbcService(PortalUserJdbcService portalUserJdbcService) {
        this.portalUserJdbcService = portalUserJdbcService;
    }

    @Resource
    public void setProjectionService(AiSensitiveWordProjectionService projectionService) {
        this.projectionService = projectionService;
    }

    @Override
    public List<AiSensitiveDetectRule> listDetectRules() {
        return aiSensitiveWordJdbcService.listDetectRules();
    }

    @Override
    public AiSensitiveDetectRule saveDetectRule(AiSensitiveDetectRule rule) {
        validateDetectRule(rule);
        AiSensitiveDetectRule saved = aiSensitiveWordJdbcService.saveDetectRule(rule);
        projectionService.syncNow();
        return saved;
    }

    @Override
    public void deleteDetectRule(Long id) {
        aiSensitiveWordJdbcService.deleteDetectRule(id);
        projectionService.syncNow();
    }

    @Override
    public List<AiSensitiveReplaceRule> listReplaceRules() {
        return aiSensitiveWordJdbcService.listReplaceRules();
    }

    @Override
    public AiSensitiveReplaceRule saveReplaceRule(AiSensitiveReplaceRule rule) {
        validateReplaceRule(rule);
        AiSensitiveReplaceRule saved = aiSensitiveWordJdbcService.saveReplaceRule(rule);
        projectionService.syncNow();
        return saved;
    }

    @Override
    public void deleteReplaceRule(Long id) {
        aiSensitiveWordJdbcService.deleteReplaceRule(id);
        projectionService.syncNow();
    }

    @Override
    public List<AiSensitiveBlockAudit> listAudits(String consumerName, String displayName, String routeName,
        String matchType, String startTime, String endTime, Integer limit) {
        return aiSensitiveWordJdbcService.listAudits(consumerName, displayName, routeName, matchType, startTime,
            endTime, limit);
    }

    @Override
    public AiSensitiveSystemConfig getSystemConfig() {
        return aiSensitiveWordJdbcService.getSystemConfig();
    }

    @Override
    public AiSensitiveSystemConfig saveSystemConfig(AiSensitiveSystemConfig config) {
        if (config == null) {
            throw new ValidationException("system config cannot be null.");
        }
        User currentUser = SessionUserHelper.getCurrentUser();
        String updatedBy = currentUser == null ? "system"
            : StringUtils.defaultIfBlank(
                StringUtils.trimToNull(currentUser.getDisplayName()),
                StringUtils.defaultIfBlank(StringUtils.trimToNull(currentUser.getName()), "system"));
        AiSensitiveSystemConfig saved = aiSensitiveWordJdbcService.saveSystemConfig(config, updatedBy);
        projectionService.syncNow();
        return saved;
    }

    @Override
    public AiSensitiveStatus getStatus() {
        return projectionService.getStatus();
    }

    @Override
    public AiSensitiveStatus reconcile() {
        projectionService.syncNow();
        return projectionService.getStatus();
    }

    @Override
    public AiSensitiveBlockAudit ingestBlockedEvent(AiSensitiveBlockAuditEvent event) {
        if (event == null) {
            throw new ValidationException("audit event cannot be null.");
        }
        String consumerName = StringUtils.trimToNull(event.getConsumerName());
        PortalUserRecord portalUser =
            consumerName == null ? null : portalUserJdbcService.queryByConsumerName(consumerName);
        return aiSensitiveWordJdbcService.saveAudit(event, portalUser == null ? null : portalUser.getDisplayName());
    }

    private void validateDetectRule(AiSensitiveDetectRule rule) {
        if (rule == null) {
            throw new ValidationException("detect rule cannot be null.");
        }
        if (StringUtils.isBlank(rule.getPattern())) {
            throw new ValidationException("detect rule pattern cannot be blank.");
        }
        String matchType = StringUtils.lowerCase(StringUtils.trimToEmpty(rule.getMatchType()));
        if (!Arrays.asList("contains", "exact", "regex").contains(matchType)) {
            throw new ValidationException("detect rule matchType must be one of contains, exact, regex.");
        }
    }

    private void validateReplaceRule(AiSensitiveReplaceRule rule) {
        if (rule == null) {
            throw new ValidationException("replace rule cannot be null.");
        }
        if (StringUtils.isBlank(rule.getPattern())) {
            throw new ValidationException("replace rule pattern cannot be blank.");
        }
        String replaceType = StringUtils.lowerCase(StringUtils.trimToEmpty(rule.getReplaceType()));
        if (!Arrays.asList("replace", "hash").contains(replaceType)) {
            throw new ValidationException("replace rule replaceType must be one of replace, hash.");
        }
    }
}
