package com.alibaba.higress.console.service;

import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.never;
import static org.mockito.Mockito.times;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import java.util.Arrays;
import java.util.Collections;
import java.util.List;

import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.mockito.ArgumentCaptor;

import com.alibaba.higress.sdk.constant.KubernetesConstants;
import com.alibaba.higress.sdk.constant.plugin.BuiltInPluginName;
import com.alibaba.higress.sdk.model.WasmPlugin;
import com.alibaba.higress.sdk.service.WasmPluginService;
import com.alibaba.higress.sdk.service.kubernetes.KubernetesClientService;
import com.alibaba.higress.sdk.service.kubernetes.KubernetesModelConverter;
import com.alibaba.higress.sdk.service.kubernetes.crd.wasm.V1alpha1WasmPlugin;
import com.alibaba.higress.sdk.service.kubernetes.crd.wasm.V1alpha1WasmPluginSpec;

import io.kubernetes.client.openapi.models.V1ObjectMeta;

class AiPluginExecutionOrderReconcileServiceTest {

    private WasmPluginService wasmPluginService;
    private KubernetesClientService kubernetesClientService;
    private KubernetesModelConverter kubernetesModelConverter;
    private AiPluginExecutionOrderReconcileService service;

    @BeforeEach
    void setUp() {
        wasmPluginService = mock(WasmPluginService.class);
        kubernetesClientService = mock(KubernetesClientService.class);
        kubernetesModelConverter = mock(KubernetesModelConverter.class);

        service = new AiPluginExecutionOrderReconcileService();
        service.setWasmPluginService(wasmPluginService);
        service.setKubernetesClientService(kubernetesClientService);
        service.setKubernetesModelConverter(kubernetesModelConverter);
    }

    @Test
    void syncNowShouldUpdateCurrentVersionOrderWhileKeepingCurrentImageSettings() throws Exception {
        WasmPlugin desiredStatistics = buildPlugin(
            BuiltInPluginName.AI_STATISTICS,
            "2.0.0",
            "STATS",
            900,
            "oci://default/ai-statistics",
            "2.0.0"
        );
        WasmPlugin desiredMasking = buildPlugin(
            BuiltInPluginName.AI_DATA_MASKING,
            "2.0.0",
            "AUTHN",
            100,
            "oci://default/ai-data-masking",
            "2.0.0"
        );

        V1alpha1WasmPlugin misorderedStatisticsCr = buildCr("ai-statistics-2.0.0", "2.0.0", "UNSPECIFIED_PHASE", 900);
        V1alpha1WasmPlugin alignedMaskingCr = buildCr("ai-data-masking-2.0.0", "2.0.0", "AUTHN", 100);

        WasmPlugin currentStatistics = buildPlugin(
            BuiltInPluginName.AI_STATISTICS,
            "2.0.0",
            "UNSPECIFIED_PHASE",
            900,
            "oci://custom/ai-statistics",
            "2.0.1"
        );

        when(wasmPluginService.query(BuiltInPluginName.AI_STATISTICS, null)).thenReturn(desiredStatistics);
        when(wasmPluginService.query(BuiltInPluginName.AI_DATA_MASKING, null)).thenReturn(desiredMasking);
        when(kubernetesClientService.listWasmPlugin(eq(BuiltInPluginName.AI_STATISTICS), eq((String) null), eq(true)))
            .thenReturn(Collections.singletonList(misorderedStatisticsCr));
        when(kubernetesClientService.listWasmPlugin(eq(BuiltInPluginName.AI_DATA_MASKING), eq((String) null), eq(true)))
            .thenReturn(Collections.singletonList(alignedMaskingCr));
        when(kubernetesModelConverter.wasmPluginFromCr(misorderedStatisticsCr)).thenReturn(currentStatistics);

        service.syncNow();

        ArgumentCaptor<WasmPlugin> pluginCaptor = ArgumentCaptor.forClass(WasmPlugin.class);
        verify(wasmPluginService).updateBuiltIn(pluginCaptor.capture());

        WasmPlugin updatedPlugin = pluginCaptor.getValue();
        Assertions.assertEquals(BuiltInPluginName.AI_STATISTICS, updatedPlugin.getName());
        Assertions.assertEquals("STATS", updatedPlugin.getPhase());
        Assertions.assertEquals(Integer.valueOf(900), updatedPlugin.getPriority());
        Assertions.assertEquals("oci://custom/ai-statistics", updatedPlugin.getImageRepository());
        Assertions.assertEquals("2.0.1", updatedPlugin.getImageVersion());
        verify(kubernetesClientService, never()).deleteWasmPlugin(any());
    }

    @Test
    void syncNowShouldMigrateLegacyBuiltInConfigToCurrentVersionAndDeleteLegacyCr() throws Exception {
        WasmPlugin desiredStatistics = buildPlugin(
            BuiltInPluginName.AI_STATISTICS,
            "2.0.0",
            "STATS",
            900,
            "oci://default/ai-statistics",
            "2.0.0"
        );
        WasmPlugin desiredMasking = buildPlugin(
            BuiltInPluginName.AI_DATA_MASKING,
            "2.0.0",
            "AUTHN",
            100,
            "oci://default/ai-data-masking",
            "2.0.0"
        );

        V1alpha1WasmPlugin legacyStatisticsCr = buildCr("ai-statistics-1.0.0", "1.0.0", "UNSPECIFIED_PHASE", 900);
        V1alpha1WasmPlugin currentStatisticsCr = buildCr("ai-statistics-2.0.0", "2.0.0", "STATS", 900);
        V1alpha1WasmPlugin currentMaskingCr = buildCr("ai-data-masking-2.0.0", "2.0.0", "AUTHN", 100);

        when(wasmPluginService.query(BuiltInPluginName.AI_STATISTICS, null)).thenReturn(desiredStatistics);
        when(wasmPluginService.query(BuiltInPluginName.AI_DATA_MASKING, null)).thenReturn(desiredMasking);
        when(kubernetesClientService.listWasmPlugin(eq(BuiltInPluginName.AI_STATISTICS), eq((String) null), eq(true)))
            .thenReturn(Arrays.asList(legacyStatisticsCr, currentStatisticsCr));
        when(kubernetesClientService.listWasmPlugin(eq(BuiltInPluginName.AI_DATA_MASKING), eq((String) null), eq(true)))
            .thenReturn(Collections.singletonList(currentMaskingCr));

        service.syncNow();

        verify(kubernetesModelConverter).mergeWasmPluginSpec(legacyStatisticsCr, currentStatisticsCr);
        verify(kubernetesClientService).replaceWasmPlugin(currentStatisticsCr);
        verify(kubernetesClientService).deleteWasmPlugin("ai-statistics-1.0.0");
        verify(wasmPluginService, never()).updateBuiltIn(any());
    }

    @Test
    void syncNowShouldSkipUpdateWhenExecutionOrderAlreadyAlignedAndNoLegacyCr() throws Exception {
        WasmPlugin desiredStatistics = buildPlugin(
            BuiltInPluginName.AI_STATISTICS,
            "2.0.0",
            "STATS",
            900,
            "oci://default/ai-statistics",
            "2.0.0"
        );
        WasmPlugin desiredMasking = buildPlugin(
            BuiltInPluginName.AI_DATA_MASKING,
            "2.0.0",
            "AUTHN",
            100,
            "oci://default/ai-data-masking",
            "2.0.0"
        );

        when(wasmPluginService.query(BuiltInPluginName.AI_STATISTICS, null)).thenReturn(desiredStatistics);
        when(wasmPluginService.query(BuiltInPluginName.AI_DATA_MASKING, null)).thenReturn(desiredMasking);
        when(kubernetesClientService.listWasmPlugin(eq(BuiltInPluginName.AI_STATISTICS), eq((String) null), eq(true)))
            .thenReturn(Collections.singletonList(buildCr("ai-statistics-2.0.0", "2.0.0", "STATS", 900)));
        when(kubernetesClientService.listWasmPlugin(eq(BuiltInPluginName.AI_DATA_MASKING), eq((String) null), eq(true)))
            .thenReturn(Collections.singletonList(buildCr("ai-data-masking-2.0.0", "2.0.0", "AUTHN", 100)));

        service.syncNow();

        verify(wasmPluginService, never()).updateBuiltIn(any());
        verify(kubernetesClientService, never()).replaceWasmPlugin(any());
        verify(kubernetesClientService, never()).deleteWasmPlugin(any());
    }

    private WasmPlugin buildPlugin(
        String name,
        String pluginVersion,
        String phase,
        Integer priority,
        String imageRepository,
        String imageVersion
    ) {
        WasmPlugin plugin = new WasmPlugin();
        plugin.setName(name);
        plugin.setPluginVersion(pluginVersion);
        plugin.setBuiltIn(true);
        plugin.setPhase(phase);
        plugin.setPriority(priority);
        plugin.setImageRepository(imageRepository);
        plugin.setImageVersion(imageVersion);
        plugin.setImagePullPolicy("UNSPECIFIED_POLICY");
        return plugin;
    }

    private V1alpha1WasmPlugin buildCr(String name, String pluginVersion, String phase, Integer priority) {
        V1alpha1WasmPlugin cr = new V1alpha1WasmPlugin();

        V1ObjectMeta metadata = new V1ObjectMeta();
        metadata.setName(name);
        metadata.putLabelsItem(KubernetesConstants.Label.WASM_PLUGIN_VERSION_KEY, pluginVersion);
        cr.setMetadata(metadata);

        V1alpha1WasmPluginSpec spec = new V1alpha1WasmPluginSpec();
        spec.setPhase(phase);
        spec.setPriority(priority);
        cr.setSpec(spec);
        return cr;
    }
}
