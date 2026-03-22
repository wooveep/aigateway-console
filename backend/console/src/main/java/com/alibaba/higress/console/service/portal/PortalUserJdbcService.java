package com.alibaba.higress.console.service.portal;

import java.sql.Connection;
import java.sql.DriverManager;
import java.sql.PreparedStatement;
import java.sql.ResultSet;
import java.sql.SQLException;
import java.sql.Timestamp;
import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

import org.apache.commons.lang3.RandomStringUtils;
import org.apache.commons.lang3.StringUtils;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.stereotype.Service;

import com.alibaba.higress.console.model.portal.PortalUserRecord;
import com.alibaba.higress.sdk.model.consumer.Consumer;

import lombok.extern.slf4j.Slf4j;

@Slf4j
@Service
public class PortalUserJdbcService {

    private static final String DEFAULT_USER_STATUS = "active";

    @Value("${higress.portal.db.url:}")
    private String dbUrl;

    @Value("${higress.portal.db.username:}")
    private String dbUsername;

    @Value("${higress.portal.db.password:}")
    private String dbPassword;

    private final BCryptPasswordEncoder passwordEncoder = new BCryptPasswordEncoder();

    public boolean enabled() {
        return StringUtils.isNotBlank(dbUrl);
    }

    public Map<String, PortalUserRecord> listByConsumerNames(List<String> consumerNames) {
        if (!enabled() || consumerNames == null || consumerNames.isEmpty()) {
            return Collections.emptyMap();
        }
        List<String> names = consumerNames.stream().filter(StringUtils::isNotBlank).distinct().collect(Collectors.toList());
        if (names.isEmpty()) {
            return Collections.emptyMap();
        }

        String placeholders = names.stream().map(i -> "?").collect(Collectors.joining(","));
        String sql = "SELECT consumer_name, display_name, email, department, status, source, last_login_at "
            + "FROM portal_user WHERE consumer_name IN (" + placeholders + ")";

        Map<String, PortalUserRecord> result = new HashMap<>();
        try (Connection connection = openConnection(); PreparedStatement statement = connection.prepareStatement(sql)) {
            for (int i = 0; i < names.size(); i++) {
                statement.setString(i + 1, names.get(i));
            }
            try (ResultSet rs = statement.executeQuery()) {
                while (rs.next()) {
                    PortalUserRecord record = mapRecord(rs);
                    result.put(record.getConsumerName(), record);
                }
            }
        } catch (SQLException ex) {
            log.warn("Failed to load portal users from MySQL.", ex);
        }
        return result;
    }

    public PortalUserRecord queryByConsumerName(String consumerName) {
        if (!enabled() || StringUtils.isBlank(consumerName)) {
            return null;
        }
        String sql = "SELECT consumer_name, display_name, email, department, status, source, last_login_at "
            + "FROM portal_user WHERE consumer_name = ?";
        try (Connection connection = openConnection(); PreparedStatement statement = connection.prepareStatement(sql)) {
            statement.setString(1, consumerName);
            try (ResultSet rs = statement.executeQuery()) {
                if (rs.next()) {
                    return mapRecord(rs);
                }
            }
        } catch (SQLException ex) {
            log.warn("Failed to query portal user {} from MySQL.", consumerName, ex);
        }
        return null;
    }

    public PortalUserRecord upsertFromConsumer(Consumer consumer, String defaultSource) {
        if (!enabled() || consumer == null || StringUtils.isBlank(consumer.getName())) {
            return null;
        }

        String consumerName = consumer.getName();
        PortalUserRecord existed = queryByConsumerName(consumerName);

        String displayName = StringUtils.firstNonBlank(consumer.getPortalDisplayName(),
            existed == null ? null : existed.getDisplayName(), consumerName);
        String email = StringUtils.firstNonBlank(consumer.getPortalEmail(), existed == null ? null : existed.getEmail(), "");
        String department = StringUtils.firstNonBlank(consumer.getDepartment(),
            existed == null ? null : existed.getDepartment(), "");
        String status = StringUtils.firstNonBlank(consumer.getPortalStatus(),
            existed == null ? null : existed.getStatus(), DEFAULT_USER_STATUS);
        String source = StringUtils.firstNonBlank(consumer.getPortalUserSource(),
            existed == null ? null : existed.getSource(), defaultSource, "console");

        String password = StringUtils.trimToNull(consumer.getPortalPassword());
        String tempPassword = null;
        if (existed == null && password == null) {
            tempPassword = RandomStringUtils.randomAlphanumeric(12);
            password = tempPassword;
        }

        try (Connection connection = openConnection()) {
            if (existed == null) {
                String insertSql = "INSERT INTO portal_user "
                    + "(consumer_name, display_name, email, department, password_hash, status, source) "
                    + "VALUES (?, ?, ?, ?, ?, ?, ?)";
                try (PreparedStatement statement = connection.prepareStatement(insertSql)) {
                    statement.setString(1, consumerName);
                    statement.setString(2, displayName);
                    statement.setString(3, email);
                    statement.setString(4, department);
                    statement.setString(5, passwordEncoder.encode(password));
                    statement.setString(6, status);
                    statement.setString(7, source);
                    statement.executeUpdate();
                }
            } else {
                String updateSql;
                if (password == null) {
                    updateSql = "UPDATE portal_user SET display_name=?, email=?, department=?, status=?, source=? "
                        + "WHERE consumer_name=?";
                } else {
                    updateSql = "UPDATE portal_user SET display_name=?, email=?, department=?, status=?, source=?, "
                        + "password_hash=? WHERE consumer_name=?";
                }
                try (PreparedStatement statement = connection.prepareStatement(updateSql)) {
                    int idx = 1;
                    statement.setString(idx++, displayName);
                    statement.setString(idx++, email);
                    statement.setString(idx++, department);
                    statement.setString(idx++, status);
                    statement.setString(idx++, source);
                    if (password != null) {
                        statement.setString(idx++, passwordEncoder.encode(password));
                    }
                    statement.setString(idx, consumerName);
                    statement.executeUpdate();
                }
            }
        } catch (SQLException ex) {
            log.warn("Failed to upsert portal user {}.", consumerName, ex);
            return null;
        }

        PortalUserRecord updated = queryByConsumerName(consumerName);
        if (updated != null) {
            updated.setTempPassword(tempPassword);
        }
        return updated;
    }

    public void updateStatus(String consumerName, String status) {
        if (!enabled() || StringUtils.isBlank(consumerName) || StringUtils.isBlank(status)) {
            return;
        }
        String sql = "UPDATE portal_user SET status = ? WHERE consumer_name = ?";
        try (Connection connection = openConnection(); PreparedStatement statement = connection.prepareStatement(sql)) {
            statement.setString(1, status);
            statement.setString(2, consumerName);
            statement.executeUpdate();
        } catch (SQLException ex) {
            log.warn("Failed to update portal user status for {}.", consumerName, ex);
        }
    }

    public void disableAllApiKeys(String consumerName) {
        if (!enabled() || StringUtils.isBlank(consumerName)) {
            return;
        }
        String sql = "UPDATE portal_api_key SET status='disabled' WHERE consumer_name = ?";
        try (Connection connection = openConnection(); PreparedStatement statement = connection.prepareStatement(sql)) {
            statement.setString(1, consumerName);
            statement.executeUpdate();
        } catch (SQLException ex) {
            log.warn("Failed to disable portal api keys for {}.", consumerName, ex);
        }
    }

    private Connection openConnection() throws SQLException {
        if (StringUtils.isBlank(dbUsername)) {
            return DriverManager.getConnection(dbUrl);
        }
        return DriverManager.getConnection(dbUrl, dbUsername, dbPassword);
    }

    private PortalUserRecord mapRecord(ResultSet rs) throws SQLException {
        Timestamp lastLogin = rs.getTimestamp("last_login_at");
        LocalDateTime lastLoginAt = null;
        if (lastLogin != null) {
            lastLoginAt = lastLogin.toLocalDateTime();
        }
        return PortalUserRecord.builder().consumerName(rs.getString("consumer_name"))
            .displayName(rs.getString("display_name")).email(rs.getString("email"))
            .department(rs.getString("department")).status(rs.getString("status"))
            .source(rs.getString("source")).lastLoginAt(lastLoginAt).build();
    }

    public List<String> listActiveRawKeys(String consumerName) {
        if (!enabled() || StringUtils.isBlank(consumerName)) {
            return Collections.emptyList();
        }
        String sql = "SELECT raw_key FROM portal_api_key WHERE consumer_name=? AND status='active' ORDER BY id ASC";
        List<String> result = new ArrayList<>();
        try (Connection connection = openConnection(); PreparedStatement statement = connection.prepareStatement(sql)) {
            statement.setString(1, consumerName);
            try (ResultSet rs = statement.executeQuery()) {
                while (rs.next()) {
                    String raw = rs.getString("raw_key");
                    if (StringUtils.isNotBlank(raw)) {
                        result.add(raw);
                    }
                }
            }
        } catch (SQLException ex) {
            log.warn("Failed to list active raw keys for {}.", consumerName, ex);
        }
        return result;
    }
}
