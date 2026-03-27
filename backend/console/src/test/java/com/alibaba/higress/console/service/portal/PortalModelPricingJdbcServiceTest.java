package com.alibaba.higress.console.service.portal;

import java.sql.Connection;
import java.sql.DriverManager;
import java.sql.PreparedStatement;
import java.sql.ResultSet;
import java.sql.SQLException;
import java.sql.Statement;
import java.util.Arrays;
import java.util.Collections;
import java.util.HashMap;
import java.util.Map;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.test.util.ReflectionTestUtils;

import com.alibaba.higress.sdk.model.ai.LlmProvider;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertTrue;

class PortalModelPricingJdbcServiceTest {

    private static final String DB_URL =
        "jdbc:h2:mem:portal_model_pricing;MODE=MySQL;DB_CLOSE_DELAY=-1;DATABASE_TO_LOWER=TRUE";

    private PortalModelPricingJdbcService service;

    @BeforeEach
    void setUp() throws Exception {
        service = new PortalModelPricingJdbcService();
        ReflectionTestUtils.setField(service, "dbUrl", DB_URL);
        ReflectionTestUtils.setField(service, "dbUsername", "sa");
        ReflectionTestUtils.setField(service, "dbPassword", "");

        try (Connection connection = openConnection(); Statement statement = connection.createStatement()) {
            statement.execute("DROP TABLE IF EXISTS billing_model_price_version");
            statement.execute("DROP TABLE IF EXISTS billing_model_catalog");
            statement.execute("CREATE TABLE billing_model_catalog ("
                + "id BIGINT AUTO_INCREMENT PRIMARY KEY,"
                + "model_id VARCHAR(128) NOT NULL UNIQUE,"
                + "name VARCHAR(128) NOT NULL,"
                + "vendor VARCHAR(128) NOT NULL,"
                + "capability VARCHAR(255) NOT NULL,"
                + "endpoint VARCHAR(255) NOT NULL,"
                + "sdk VARCHAR(128) NOT NULL,"
                + "summary CLOB NOT NULL,"
                + "status VARCHAR(16) NOT NULL)");
            statement.execute("CREATE TABLE billing_model_price_version ("
                + "id BIGINT AUTO_INCREMENT PRIMARY KEY,"
                + "model_id VARCHAR(128) NOT NULL,"
                + "currency VARCHAR(8) NOT NULL,"
                + "input_price_per_1k_micro_yuan BIGINT NOT NULL,"
                + "output_price_per_1k_micro_yuan BIGINT NOT NULL,"
                + "effective_from TIMESTAMP NOT NULL,"
                + "effective_to TIMESTAMP NULL,"
                + "status VARCHAR(16) NOT NULL)");
        }
    }

    @Test
    void upsertProviderWritesActiveCatalogAndPriceVersion() throws Exception {
        service.upsertProvider(providerWithPricing(0.0025D, 0.0065D));

        try (Connection connection = openConnection()) {
            try (PreparedStatement statement = connection.prepareStatement(
                "SELECT model_id, vendor, endpoint, sdk, summary, status "
                    + "FROM billing_model_catalog WHERE model_id = ?")) {
                statement.setString(1, "qwen-plus");
                try (ResultSet rs = statement.executeQuery()) {
                    assertTrue(rs.next());
                    assertEquals("qwen-plus", rs.getString("model_id"));
                    assertEquals("aliyun", rs.getString("vendor"));
                    assertEquals("https://dashscope.aliyuncs.com/compatible-mode/v1", rs.getString("endpoint"));
                    assertEquals("openai", rs.getString("sdk"));
                    assertEquals("Tongyi Qwen Plus", rs.getString("summary"));
                    assertEquals("active", rs.getString("status"));
                }
            }

            try (PreparedStatement statement = connection.prepareStatement(
                "SELECT currency, input_price_per_1k_micro_yuan, output_price_per_1k_micro_yuan, "
                    + "effective_to, status "
                    + "FROM billing_model_price_version WHERE model_id = ?")) {
                statement.setString(1, "qwen-plus");
                try (ResultSet rs = statement.executeQuery()) {
                    assertTrue(rs.next());
                    assertEquals("CNY", rs.getString("currency"));
                    assertEquals(2500L, rs.getLong("input_price_per_1k_micro_yuan"));
                    assertEquals(6500L, rs.getLong("output_price_per_1k_micro_yuan"));
                    assertEquals("active", rs.getString("status"));
                    assertEquals(null, rs.getTimestamp("effective_to"));
                }
            }
        }
    }

    @Test
    void upsertProviderRotatesActivePriceVersionWhenPricingChanges() throws Exception {
        service.upsertProvider(providerWithPricing(0.0025D, 0.0065D));
        service.upsertProvider(providerWithPricing(0.0030D, 0.0070D));

        try (Connection connection = openConnection()) {
            try (PreparedStatement countStatement = connection.prepareStatement(
                "SELECT COUNT(1) FROM billing_model_price_version WHERE model_id = ?")) {
                countStatement.setString(1, "qwen-plus");
                try (ResultSet rs = countStatement.executeQuery()) {
                    assertTrue(rs.next());
                    assertEquals(2, rs.getInt(1));
                }
            }

            try (PreparedStatement activeStatement = connection.prepareStatement(
                "SELECT input_price_per_1k_micro_yuan, output_price_per_1k_micro_yuan, effective_to, status "
                    + "FROM billing_model_price_version "
                    + "WHERE model_id = ? AND effective_to IS NULL")) {
                activeStatement.setString(1, "qwen-plus");
                try (ResultSet rs = activeStatement.executeQuery()) {
                    assertTrue(rs.next());
                    assertEquals(3000L, rs.getLong("input_price_per_1k_micro_yuan"));
                    assertEquals(7000L, rs.getLong("output_price_per_1k_micro_yuan"));
                    assertEquals("active", rs.getString("status"));
                }
            }

            try (PreparedStatement inactiveStatement = connection.prepareStatement(
                "SELECT effective_to, status "
                    + "FROM billing_model_price_version "
                    + "WHERE model_id = ? AND effective_to IS NOT NULL")) {
                inactiveStatement.setString(1, "qwen-plus");
                try (ResultSet rs = inactiveStatement.executeQuery()) {
                    assertTrue(rs.next());
                    assertNotNull(rs.getTimestamp("effective_to"));
                    assertEquals("inactive", rs.getString("status"));
                }
            }
        }
    }

    private Connection openConnection() throws SQLException {
        return DriverManager.getConnection(DB_URL, "sa", "");
    }

    private LlmProvider providerWithPricing(double inputPer1K, double outputPer1K) {
        Map<String, Object> pricing = new HashMap<>();
        pricing.put("currency", "CNY");
        pricing.put("inputPer1K", inputPer1K);
        pricing.put("outputPer1K", outputPer1K);

        Map<String, Object> capabilities = new HashMap<>();
        capabilities.put("modalities", Arrays.asList("text"));
        capabilities.put("features", Arrays.asList("tool_call"));

        Map<String, Object> portalModelMeta = new HashMap<>();
        portalModelMeta.put("intro", "Tongyi Qwen Plus");
        portalModelMeta.put("capabilities", capabilities);
        portalModelMeta.put("pricing", pricing);

        Map<String, Object> rawConfigs = new HashMap<>();
        rawConfigs.put("openaiCustomUrl", "https://dashscope.aliyuncs.com/compatible-mode/v1");
        rawConfigs.put("portalModelMeta", portalModelMeta);

        return LlmProvider.builder()
            .name("qwen-plus")
            .type("aliyun")
            .protocol("openai")
            .tokens(Collections.singletonList("token-a"))
            .rawConfigs(rawConfigs)
            .build();
    }
}
