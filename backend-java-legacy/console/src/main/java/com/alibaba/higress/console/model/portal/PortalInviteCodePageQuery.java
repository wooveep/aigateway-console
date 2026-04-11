package com.alibaba.higress.console.model.portal;

import com.alibaba.higress.sdk.model.CommonPageQuery;

import io.swagger.v3.oas.annotations.media.Schema;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.EqualsAndHashCode;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
@EqualsAndHashCode(callSuper = true)
@Schema(description = "Query criteria for portal invite code listing.")
public class PortalInviteCodePageQuery extends CommonPageQuery {

    @Schema(description = "Invite code status filter, e.g. active/disabled.")
    private String status;
}
