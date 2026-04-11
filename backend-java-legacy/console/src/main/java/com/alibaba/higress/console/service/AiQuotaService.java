package com.alibaba.higress.console.service;

import java.util.List;

import com.alibaba.higress.console.model.aiquota.AiQuotaConsumerQuota;
import com.alibaba.higress.console.model.aiquota.AiQuotaMenuState;
import com.alibaba.higress.console.model.aiquota.AiQuotaRouteSummary;
import com.alibaba.higress.console.model.aiquota.AiQuotaScheduleRule;
import com.alibaba.higress.console.model.aiquota.AiQuotaScheduleRuleRequest;
import com.alibaba.higress.console.model.aiquota.AiQuotaUserPolicy;
import com.alibaba.higress.console.model.aiquota.AiQuotaUserPolicyRequest;

public interface AiQuotaService {

    AiQuotaMenuState getMenuState();

    List<AiQuotaRouteSummary> listEnabledRoutes();

    List<AiQuotaConsumerQuota> listConsumerQuotas(String routeName);

    AiQuotaConsumerQuota refreshQuota(String routeName, String consumerName, long quota);

    AiQuotaConsumerQuota deltaQuota(String routeName, String consumerName, long delta);

    AiQuotaUserPolicy getUserPolicy(String routeName, String consumerName);

    AiQuotaUserPolicy saveUserPolicy(String routeName, String consumerName, AiQuotaUserPolicyRequest request);

    List<AiQuotaScheduleRule> listScheduleRules(String routeName, String consumerName);

    AiQuotaScheduleRule saveScheduleRule(String routeName, AiQuotaScheduleRuleRequest request);

    void deleteScheduleRule(String routeName, String ruleId);
}
