package com.alibaba.higress.console.service.portal;

import java.math.BigDecimal;
import java.math.RoundingMode;
import java.sql.Connection;
import java.sql.DriverManager;
import java.sql.PreparedStatement;
import java.sql.ResultSet;
import java.sql.SQLException;
import java.sql.Timestamp;
import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;

import javax.annotation.PostConstruct;

import org.apache.commons.lang3.StringUtils;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import com.alibaba.higress.sdk.exception.BusinessException;
import com.alibaba.higress.sdk.exception.ValidationException;
import com.alibaba.higress.sdk.model.ai.LlmProvider;

import lombok.extern.slf4j.Slf4j;

@Slf4j
@Service
public class PortalModelPricingJdbcService {

    private static final String PORTAL_MODEL_META_KEY = "portalModelMeta";
    private static final String PRICING_KEY = "pricing";
    private static final String CAPABILITIES_KEY = "capabilities";
    private static final String CURRENCY_CNY = "CNY";
    private static final String STATUS_ACTIVE = "active";
    private static final String STATUS_INACTIVE = "inactive";
    private static final String STATUS_DISABLED = "disabled";
    private static final long MICRO_YUAN_PER_RMB = 1_000_000L;
    private static final String DEFAULT_ENDPOINT = "-";
    private static final String DEFAULT_SDK = "openai/v1";

    @Value("${higress.portal.db.url:}")
    private String dbUrl;

    @Value("${higress.portal.db.username:}")
    private String dbUsername;

    @Value("${higress.portal.db.password:}")
    private String dbPassword;

    @PostConstruct
    public void init() {
        ensureBillingModelTables();
    }

    public boolean enabled() {
        return StringUtils.isNotBlank(dbUrl);
    }

    public void upsertProvider(LlmProvider provider) {
        if (!enabled()) {
            throw new IllegalStateException("Portal database is unavailable.");
        }
        ProviderModelPricingMeta meta = extractMeta(provider);
        try (Connection connection = openConnection()) {
            connection.setAutoCommit(false);
            try {
                upsertCatalog(connection, meta);
                upsertPriceVersion(connection, meta);
                connection.commit();
            } catch (SQLException ex) {
                connection.rollback();
                throw ex;
            } finally {
                connection.setAutoCommit(true);
            }
        } catch (SQLException ex) {
            log.warn("Failed to sync provider {} pricing into Portal MySQL.", provider == null ? null : provider.getName(), ex);
            throw new BusinessException("Failed to sync provider pricing into Portal.", ex);
        }
    }

    public void disableProvider(String providerName) {
        if (!enabled()) {
            throw new IllegalStateException("Portal database is unavailable.");
        }
        String normalizedName = StringUtils.trimToNull(providerName);
        if (normalizedName == null) {
            throw new ValidationException("providerName cannot be blank.");
        }

        try (Connection connection = openConnection()) {
            connection.setAutoCommit(false);
            try (PreparedStatement catalogStmt = connection.prepareStatement(
                "UPDATE billing_model_catalog SET status = ? WHERE model_id = ?");
                PreparedStatement versionStmt = connection.prepareStatement(
                    "UPDATE billing_model_price_version SET status = ?, effective_to = ? "
                        + "WHERE model_id = ? AND effective_to IS NULL")) {
                catalogStmt.setString(1, STATUS_DISABLED);
                catalogStmt.setString(2, normalizedName);
                catalogStmt.executeUpdate();

                versionStmt.setString(1, STATUS_INACTIVE);
                versionStmt.setTimestamp(2, Timestamp.valueOf(LocalDateTime.now()));
                versionStmt.setString(3, normalizedName);
                versionStmt.executeUpdate();

                connection.commit();
            } catch (SQLException ex) {
                connection.rollback();
                throw ex;
            } finally {
                connection.setAutoCommit(true);
            }
        } catch (SQLException ex) {
            log.warn("Failed to disable Portal model pricing for provider {}.", normalizedName, ex);
            throw new BusinessException("Failed to disable provider pricing in Portal.", ex);
        }
    }

    private void ensureBillingModelTables() {
        if (!enabled()) {
            return;
        }
        String[] ddls = new String[] {
            "CREATE TABLE IF NOT EXISTS billing_model_catalog ("
                + "id BIGINT AUTO_INCREMENT PRIMARY KEY,"
                + "model_id VARCHAR(128) NOT NULL UNIQUE,"
                + "name VARCHAR(128) NOT NULL,"
                + "vendor VARCHAR(128) NOT NULL,"
                + "capability VARCHAR(255) NOT NULL,"
                + "endpoint VARCHAR(255) NOT NULL,"
                + "sdk VARCHAR(128) NOT NULL,"
                + "summary TEXT NOT NULL,"
                + "status VARCHAR(16) NOT NULL DEFAULT 'active',"
                + "created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,"
                + "updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,"
                + "INDEX idx_billing_model_status (status)"
                + ") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4",
            "CREATE TABLE IF NOT EXISTS billing_model_price_version ("
                + "id BIGINT AUTO_INCREMENT PRIMARY KEY,"
                + "model_id VARCHAR(128) NOT NULL,"
                + "currency VARCHAR(8) NOT NULL DEFAULT 'CNY',"
                + "input_price_per_1k_micro_yuan BIGINT NOT NULL DEFAULT 0,"
                + "output_price_per_1k_micro_yuan BIGINT NOT NULL DEFAULT 0,"
                + "effective_from DATETIME NOT NULL,"
                + "effective_to DATETIME NULL,"
                + "status VARCHAR(16) NOT NULL DEFAULT 'active',"
                + "created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,"
                + "updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,"
                + "INDEX idx_billing_model_price_active (model_id, status, effective_to),"
                + "INDEX idx_billing_model_price_time (effective_from)"
                + ") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4",
        };

        try (Connection connection = openConnection()) {
            for (String ddl : ddls) {
                try (PreparedStatement statement = connection.prepareStatement(ddl)) {
                    statement.execute();
                }
            }
        } catch (SQLException ex) {
            log.warn("Failed to ensure Portal billing model pricing tables.", ex);
        }
    }

    private void upsertCatalog(Connection connection, ProviderModelPricingMeta meta) throws SQLException {
        String sql = "INSERT INTO billing_model_catalog "
            + "(model_id, name, vendor, capability, endpoint, sdk, summary, status) "
            + "VALUES (?, ?, ?, ?, ?, ?, ?, ?) "
            + "ON DUPLICATE KEY UPDATE "
            + "name = VALUES(name), "
            + "vendor = VALUES(vendor), "
            + "capability = VALUES(capability), "
            + "endpoint = VALUES(endpoint), "
            + "sdk = VALUES(sdk), "
            + "summary = VALUES(summary), "
            + "status = VALUES(status)";
        try (PreparedStatement statement = connection.prepareStatement(sql)) {
            statement.setString(1, meta.getModelID());
            statement.setString(2, meta.getName());
            statement.setString(3, meta.getVendor());
            statement.setString(4, meta.getCapability());
            statement.setString(5, meta.getEndpoint());
            statement.setString(6, meta.getSdk());
            statement.setString(7, meta.getSummary());
            statement.setString(8, STATUS_ACTIVE);
            statement.executeUpdate();
        }
    }

    private void upsertPriceVersion(Connection connection, ProviderModelPricingMeta meta) throws SQLException {
        String selectSql = "SELECT id, currency, input_price_per_1k_micro_yuan, output_price_per_1k_micro_yuan "
            + "FROM billing_model_price_version "
            + "WHERE model_id = ? AND effective_to IS NULL "
            + "ORDER BY id DESC LIMIT 1";

        PriceVersionState current = null;
        try (PreparedStatement statement = connection.prepareStatement(selectSql)) {
            statement.setString(1, meta.getModelID());
            try (ResultSet rs = statement.executeQuery()) {
                if (rs.next()) {
                    current = new PriceVersionState(rs.getLong("id"), rs.getString("currency"),
                        rs.getLong("input_price_per_1k_micro_yuan"), rs.getLong("output_price_per_1k_micro_yuan"));
                }
            }
        }

        if (current != null && StringUtils.equalsIgnoreCase(meta.getCurrency(), current.getCurrency())
            && meta.getInputPricePer1KMicroYuan() == current.getInputPricePer1KMicroYuan()
            && meta.getOutputPricePer1KMicroYuan() == current.getOutputPricePer1KMicroYuan()) {
            try (PreparedStatement statement = connection.prepareStatement(
                "UPDATE billing_model_price_version SET status = ?, effective_to = NULL WHERE id = ?")) {
                statement.setString(1, STATUS_ACTIVE);
                statement.setLong(2, current.getId());
                statement.executeUpdate();
            }
            return;
        }

        LocalDateTime now = LocalDateTime.now();
        try (PreparedStatement deactivate = connection.prepareStatement(
            "UPDATE billing_model_price_version SET status = ?, effective_to = ? "
                + "WHERE model_id = ? AND effective_to IS NULL");
            PreparedStatement insert = connection.prepareStatement(
                "INSERT INTO billing_model_price_version "
                    + "(model_id, currency, input_price_per_1k_micro_yuan, output_price_per_1k_micro_yuan, "
                    + "effective_from, status) VALUES (?, ?, ?, ?, ?, ?)")) {
            deactivate.setString(1, STATUS_INACTIVE);
            deactivate.setTimestamp(2, Timestamp.valueOf(now));
            deactivate.setString(3, meta.getModelID());
            deactivate.executeUpdate();

            insert.setString(1, meta.getModelID());
            insert.setString(2, meta.getCurrency());
            insert.setLong(3, meta.getInputPricePer1KMicroYuan());
            insert.setLong(4, meta.getOutputPricePer1KMicroYuan());
            insert.setTimestamp(5, Timestamp.valueOf(now));
            insert.setString(6, STATUS_ACTIVE);
            insert.executeUpdate();
        }
    }

    private ProviderModelPricingMeta extractMeta(LlmProvider provider) {
        if (provider == null || StringUtils.isBlank(provider.getName())) {
            throw new ValidationException("provider name cannot be blank.");
        }
        Map<String, Object> rawConfigs = provider.getRawConfigs();
        if (rawConfigs == null || rawConfigs.isEmpty()) {
            throw new ValidationException("rawConfigs.portalModelMeta is required.");
        }

        Map<String, Object> portalModelMeta = requireMap(rawConfigs.get(PORTAL_MODEL_META_KEY),
            "rawConfigs.portalModelMeta");
        Map<String, Object> pricing = requireMap(portalModelMeta.get(PRICING_KEY),
            "rawConfigs.portalModelMeta.pricing");

        String currency = StringUtils.upperCase(StringUtils.defaultIfBlank(asString(pricing.get("currency")), CURRENCY_CNY));
        if (!CURRENCY_CNY.equals(currency)) {
            throw new ValidationException("rawConfigs.portalModelMeta.pricing.currency must be CNY.");
        }

        long inputPricePer1KMicroYuan = toMicroYuan(requireNumber(pricing.get("inputPer1K"),
            "rawConfigs.portalModelMeta.pricing.inputPer1K"));
        long outputPricePer1KMicroYuan = toMicroYuan(requireNumber(pricing.get("outputPer1K"),
            "rawConfigs.portalModelMeta.pricing.outputPer1K"));

        String intro = StringUtils.trimToEmpty(asString(portalModelMeta.get("intro")));
        String vendor = StringUtils.defaultIfBlank(StringUtils.trimToEmpty(provider.getType()), "unknown");
        String capability = buildCapabilitySummary(portalModelMeta, intro, vendor);
        String summary = StringUtils.defaultIfBlank(intro, capability);

        return ProviderModelPricingMeta.builder()
            .modelID(StringUtils.trim(provider.getName()))
            .name(StringUtils.trim(provider.getName()))
            .vendor(vendor)
            .capability(capability)
            .endpoint(resolveEndpoint(rawConfigs))
            .sdk(StringUtils.defaultIfBlank(StringUtils.trimToEmpty(provider.getProtocol()), DEFAULT_SDK))
            .summary(summary)
            .currency(currency)
            .inputPricePer1KMicroYuan(inputPricePer1KMicroYuan)
            .outputPricePer1KMicroYuan(outputPricePer1KMicroYuan)
            .build();
    }

    private String buildCapabilitySummary(Map<String, Object> portalModelMeta, String intro, String vendor) {
        Object capabilitiesObj = portalModelMeta.get(CAPABILITIES_KEY);
        if (!(capabilitiesObj instanceof Map)) {
            return StringUtils.defaultIfBlank(intro, vendor);
        }
        @SuppressWarnings("unchecked")
        Map<String, Object> capabilities = (Map<String, Object>)capabilitiesObj;
        List<String> parts = new ArrayList<>();
        parts.addAll(readStringList(capabilities.get("modalities")));
        parts.addAll(readStringList(capabilities.get("features")));
        String combined = StringUtils.join(parts, " / ");
        return StringUtils.defaultIfBlank(combined, StringUtils.defaultIfBlank(intro, vendor));
    }

    private List<String> readStringList(Object value) {
        List<String> result = new ArrayList<>();
        if (!(value instanceof List)) {
            return result;
        }
        @SuppressWarnings("unchecked")
        List<Object> values = (List<Object>)value;
        for (Object item : values) {
            String text = StringUtils.trimToNull(asString(item));
            if (text != null) {
                result.add(text);
            }
        }
        return result;
    }

    private String resolveEndpoint(Map<String, Object> rawConfigs) {
        String[] candidateKeys = new String[] {
            "openaiCustomUrl",
            "azureServiceUrl",
            "qwenDomain",
            "zhipuDomain",
            "ollamaServerHost",
        };
        for (String key : candidateKeys) {
            String value = StringUtils.trimToNull(asString(rawConfigs.get(key)));
            if (value != null) {
                return value;
            }
        }
        return DEFAULT_ENDPOINT;
    }

    private double requireNumber(Object value, String path) {
        Double numberValue = parseNumber(value);
        if (numberValue == null) {
            throw new ValidationException(path + " must be a number.");
        }
        if (numberValue < 0) {
            throw new ValidationException(path + " cannot be negative.");
        }
        return numberValue;
    }

    private Double parseNumber(Object value) {
        if (value instanceof Number) {
            return ((Number)value).doubleValue();
        }
        if (value instanceof String) {
            String text = StringUtils.trimToNull((String)value);
            if (text == null) {
                return null;
            }
            try {
                return Double.parseDouble(text);
            } catch (NumberFormatException ex) {
                return null;
            }
        }
        return null;
    }

    private Map<String, Object> requireMap(Object value, String path) {
        if (!(value instanceof Map)) {
            throw new ValidationException(path + " must be an object.");
        }
        @SuppressWarnings("unchecked")
        Map<String, Object> result = (Map<String, Object>)value;
        return result;
    }

    private String asString(Object value) {
        return value instanceof String ? (String)value : null;
    }

    private long toMicroYuan(double amount) {
        return BigDecimal.valueOf(amount).multiply(BigDecimal.valueOf(MICRO_YUAN_PER_RMB))
            .setScale(0, RoundingMode.HALF_UP).longValue();
    }

    private Connection openConnection() throws SQLException {
        if (StringUtils.isBlank(dbUsername)) {
            return DriverManager.getConnection(dbUrl);
        }
        return DriverManager.getConnection(dbUrl, dbUsername, dbPassword);
    }

    @lombok.Data
    @lombok.Builder
    private static class ProviderModelPricingMeta {
        private String modelID;
        private String name;
        private String vendor;
        private String capability;
        private String endpoint;
        private String sdk;
        private String summary;
        private String currency;
        private long inputPricePer1KMicroYuan;
        private long outputPricePer1KMicroYuan;
    }

    @lombok.AllArgsConstructor
    @lombok.Getter
    private static class PriceVersionState {
        private final long id;
        private final String currency;
        private final long inputPricePer1KMicroYuan;
        private final long outputPricePer1KMicroYuan;
    }
}
