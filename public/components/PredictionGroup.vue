<!--

    Copyright © 2021 Uncharted Software Inc.

    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at

        http://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.
-->
<template>
  <div v-if="isOpen" class="prediction-group">
    <component
      :is="getFacetByType(predictedSummary.type)"
      :key="predictedSummary.key"
      enable-highlighting
      :summary="predictedSummary"
      :geo-enabled="hasGeoData && isActivePrediction"
      :highlights="highlights"
      :enabled-type-changes="[]"
      :row-selection="rowSelection"
      :instance-name="predictionInstanceName"
      @facet-click="onCategoricalClick"
      @numerical-click="onNumericalClick"
      @range-change="onRangeChange"
    />
    <component
      :is="getFacetByType(confidenceSummary.type)"
      v-if="!!confidenceSummary"
      :key="confidenceSummary.key"
      enable-highlighting
      :summary="confidenceSummary"
      :geo-enabled="hasGeoData && isActivePrediction"
      :highlights="highlights"
      :enabled-type-changes="[]"
      :row-selection="rowSelection"
      :instance-name="confidenceInstanceName"
      @facet-click="onCategoricalClick"
      @numerical-click="onNumericalClick"
      @range-change="onRangeChange"
    />
    <component
      :is="getFacetByType(rankingSummary.type)"
      v-if="!!rankingSummary"
      :key="rankingSummary.key"
      enable-highlighting
      :geo-enabled="hasGeoData && isActivePrediction"
      :summary="rankingSummary"
      :highlights="highlights"
      :enabled-type-changes="[]"
      :row-selection="rowSelection"
      :instance-name="rankingInstanceName"
      @facet-click="onCategoricalClick"
      @numerical-click="onNumericalClick"
      @range-change="onRangeChange"
    />
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import FacetNumerical from "../components/facets/FacetNumerical.vue";
import FacetCategorical from "../components/facets/FacetCategorical.vue";
import {
  VariableSummary,
  RowSelection,
  Highlight,
} from "../store/dataset/index";
import { getters as routeGetters } from "../store/route/module";
import { getters as requestGetters } from "../store/requests/module";
import { getFacetByType } from "../util/facets";

import { Predictions } from "../store/requests";
import { isTopPredictionByTime } from "../util/predictions";
import { reviseOpenSolutions } from "../util/solutions";
import { EventList } from "../util/events";
export default Vue.extend({
  name: "PredictionGroup",
  components: {
    FacetNumerical,
    FacetCategorical,
  },
  props: {
    confidenceSummary: Object as () => VariableSummary,
    predictedSummary: Object as () => VariableSummary,
    rankingSummary: Object as () => VariableSummary,
    highlights: Array as () => Highlight[],
    prediction: Object as () => Predictions,
  },
  computed: {
    isOpen(): boolean {
      return this.openSolution.has(this.prediction?.requestId);
    },
    hasGeoData(): boolean {
      return routeGetters.hasGeoData(this.$store);
    },
    rowSelection(): RowSelection {
      return routeGetters.getDecodedRowSelection(this.$store);
    },
    confidenceInstanceName(): string {
      return `confidence-prediction-facet-${this.predictionDataset}`;
    },
    predictionDataset(): string {
      return routeGetters.getRoutePredictionsDataset(this.$store);
    },
    predictionInstanceName(): string {
      return `prediction-facet-${this.predictionDataset}`;
    },
    rankingInstanceName(): string {
      return `ranking-prediction-facet-${this.predictionDataset}`;
    },
    isActivePrediction(): boolean {
      return (
        routeGetters.getRouteProduceRequestId(this.$store) ===
        this.prediction?.requestId
      );
    },
    openSolution(): Map<string, boolean> {
      return new Map(
        routeGetters.getRouteOpenSolutions(this.$store).map((s) => {
          return [s, true];
        })
      );
    },
    isTopN(): boolean {
      return isTopPredictionByTime(
        requestGetters.getRelevantPredictions(this.$store),
        this.prediction?.requestId,
        3
      );
    },
    isOpenInRoute(): boolean {
      return this.openPredictions.some((s) => {
        return s === this.prediction?.requestId;
      });
    },
    openPredictions(): string[] {
      return routeGetters.getRouteOpenSolutions(this.$store);
    },
  },
  mounted() {
    if (
      (this.isActivePrediction && !this.isOpenInRoute) ||
      (this.isTopN && !this.isOpenInRoute && this.openPredictions.length < 3)
    ) {
      reviseOpenSolutions(
        this.prediction?.requestId,
        this.$route,
        this.$router
      );
    }
  },
  methods: {
    getFacetByType(type: string) {
      return getFacetByType(type);
    },
    onCategoricalClick(
      context: string,
      key: string,
      value: string,
      dataset: string
    ) {
      this.$emit(EventList.FACETS.CATEGORICAL_CLICK_EVENT, {
        context,
        key,
        value,
        dataset,
      });
    },

    onNumericalClick(
      context: string,
      key: string,
      value: { from: number; to: number },
      dataset: string
    ) {
      this.$emit(EventList.FACETS.NUMERICAL_CLICK_EVENT, {
        context,
        key,
        value,
        dataset,
      });
    },

    onRangeChange(
      context: string,
      key: string,
      value: { from: { label: string[] }; to: { label: string[] } },
      dataset: string
    ) {
      this.$emit(EventList.FACETS.RANGE_CHANGE_EVENT, {
        context,
        key,
        value,
        dataset,
      });
    },
  },
});
</script>
<style scoped>
.prediction-group {
  margin: 5px;
  padding: 10px;
  border-bottom-style: solid;
  border-bottom-color: lightgray;
  border-bottom-width: 1px;
}

.prediction-group-title {
  vertical-align: middle;
}

.prediction-group-title .badge {
  display: inline;
  vertical-align: middle;
  padding: 0.45em 0.4em 0.3em 0.4em;
}

.prediction-group-body {
  padding: 4px 0;
}
</style>
