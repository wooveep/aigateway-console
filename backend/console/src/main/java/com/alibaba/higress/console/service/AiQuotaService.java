package com.alibaba.higress.console.service;

import java.util.List;

import com.alibaba.higress.console.model.aiquota.AiQuotaConsumerQuota;
import com.alibaba.higress.console.model.aiquota.AiQuotaMenuState;
import com.alibaba.higress.console.model.aiquota.AiQuotaRouteSummary;
import com.alibaba.higress.console.model.aiquota.AiQuotaScheduleRule;
import com.alibaba.higress.console.model.aiquota.AiQuotaScheduleRuleRequest;

public interface AiQuotaService {

    AiQuotaMenuState getMenuState();

    List<AiQuotaRouteSummary> listEnabledRoutes();

    List<AiQuotaConsumerQuota> listConsumerQuotas(String routeName);

    AiQuotaConsumerQuota refreshQuota(String routeName, String consumerName, int quota);

    AiQuotaConsumerQuota deltaQuota(String routeName, String consumerName, int delta);

    List<AiQuotaScheduleRule> listScheduleRules(String routeName, String consumerName);

    AiQuotaScheduleRule saveScheduleRule(String routeName, AiQuotaScheduleRuleRequest request);

    void deleteScheduleRule(String routeName, String ruleId);
}
