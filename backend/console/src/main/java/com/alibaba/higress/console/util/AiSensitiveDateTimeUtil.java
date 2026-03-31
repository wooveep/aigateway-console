package com.alibaba.higress.console.util;

import java.sql.Timestamp;
import java.time.Instant;
import java.time.LocalDateTime;
import java.time.OffsetDateTime;
import java.time.ZoneId;
import java.time.format.DateTimeFormatter;
import java.time.format.DateTimeParseException;

import org.apache.commons.lang3.StringUtils;

import com.alibaba.higress.sdk.exception.ValidationException;

public final class AiSensitiveDateTimeUtil {

    private static final DateTimeFormatter LOCAL_DATE_TIME_MINUTE_FORMATTER =
        DateTimeFormatter.ofPattern("yyyy-MM-dd'T'HH:mm");

    private AiSensitiveDateTimeUtil() {
    }

    public static String formatTimestamp(Timestamp timestamp) {
        return timestamp == null ? null : timestamp.toInstant().toString();
    }

    public static String formatLocalDateTime(LocalDateTime value) {
        return value == null ? null : value.atZone(ZoneId.systemDefault()).toInstant().toString();
    }

    public static Timestamp parseTimestamp(String value, String fieldName) {
        String normalized = StringUtils.trimToEmpty(value);
        if (normalized.isEmpty()) {
            return null;
        }
        try {
            return Timestamp.from(Instant.parse(normalized));
        } catch (DateTimeParseException ignored) {
        }
        try {
            return Timestamp.from(OffsetDateTime.parse(normalized).toInstant());
        } catch (DateTimeParseException ignored) {
        }
        try {
            return Timestamp.from(LocalDateTime.parse(normalized, DateTimeFormatter.ISO_LOCAL_DATE_TIME)
                .atZone(ZoneId.systemDefault()).toInstant());
        } catch (DateTimeParseException ignored) {
        }
        try {
            return Timestamp.from(LocalDateTime.parse(normalized, LOCAL_DATE_TIME_MINUTE_FORMATTER)
                .atZone(ZoneId.systemDefault()).toInstant());
        } catch (DateTimeParseException ex) {
            throw new ValidationException(fieldName + " must be RFC3339 or yyyy-MM-dd'T'HH:mm[:ss].");
        }
    }
}
