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
  <div class="w-100 nav-bar-margin d-flex flex-column">
    <header ref="header" class="bg-white">
      <search-bar
        v-show="!isBusy"
        :variables="allVariables"
        :filters="filters"
        :highlights="routeHighlight"
        handle-updates
      />
    </header>
    <div class="view-container">
      <action-column
        ref="action-column"
        :actions="activeActions"
        :current-action="currentAction"
        @set-active-pane="onSetActive"
      />
      <left-side-panel v-if="currentAction !== ''" :panel-title="currentAction">
        <add-variable-pane
          v-if="activePane === 'add'"
          :enable-label="imageVarExists"
          @label="switchToLabelState"
        />
        <export-pane v-else-if="activePane === 'export'" />
        <template v-else>
          <template v-if="hasNoVariables">
            <p v-if="activePane === 'selected'">
              Select a variable to explore.
            </p>
            <p v-else>All the variables of that type are selected.</p>
          </template>
          <facet-list-pane
            v-else
            :enable-explore="state.name === 'select'"
            :enable-training-target="state.name === 'select'"
            :is-target-panel="activePane === 'target' && isSelectState"
            :variables="activeVariables"
            :enable-color-scales="geoVarExists"
            :include="include"
            :summaries="summaries"
            :enable-footer="isSelectState"
            :dataset="dataset"
            :instance-name="activePane"
            @fetch-summaries="fetchSummaries"
            @type-change="fetchSummaries"
          />
        </template>
      </left-side-panel>
      <main class="content">
        <loading-spinner v-show="isBusy" :state="busyState" />
        <template>
          <!-- Tabs to switch views -->

          <div v-if="!isBusy" class="d-flex flex-row align-items-center mt-2">
            <div class="flex-grow-1 mr-2">
              <b-tabs v-model="activeView" class="tab-container">
                <b-tab
                  v-for="(view, index) in activeViews"
                  :key="index"
                  :active="view === activeViews[activeView]"
                  :title="capitalize(view)"
                  @click="onTabClick(view)"
                >
                  <template v-slot:title>
                    <b-spinner v-if="dataLoading" small />
                  </template>
                </b-tab>
              </b-tabs>
            </div>
            <color-scale-selection
              v-if="isMultiBandImage"
              class="align-self-center mr-2"
            />
            <layer-selection
              v-if="isMultiBandImage"
              :has-image-attention="isResultState"
              class="align-self-center mr-2"
            />
            <b-button
              v-if="include && isSelectState"
              class="select-data-action-exclude align-self-center"
              variant="outline-secondary"
              :disabled="isExcludeDisabled"
              @click="onExcludeClick"
            >
              <i
                class="fa fa-minus-circle pr-1"
                :class="{
                  'exclude-highlight': isFilteringHighlights,
                  'exclude-selection': isFilteringSelection,
                }"
              />
              Exclude
            </b-button>
            <label-header-buttons v-if="isLabelState" class="height-36" />
            <legend-weight
              v-if="hasWeight && isResultState"
              class="ml-5 mr-auto"
            />
          </div>
          <section v-show="!isBusy" class="data-container">
            <component
              :is="viewComponent"
              ref="dataView"
              :instance-name="instanceName"
              :included-active="include"
              :dataset="dataset"
              :data-fields="fields"
              :timeseries-info="timeseries"
              :data-items="items"
              :item-count="items.length"
              :baseline-items="baselineItems"
              :baseline-map="baselineMap"
              :summaries="summaries"
              :solution="solution"
              :residual-extrema="residualExtrema"
              :enable-selection-tool-event="isLabelState"
              :variables="allVariables"
              :label-feature-name="labelName"
              :label-score-name="labelName"
              :area-of-interest-items="{
                inner: drillDownBaseline,
                outer: drillDownFiltered,
              }"
              :get-timeseries="state.getTimeseries"
              @tile-clicked="onTileClick"
              @fetch-timeseries="fetchTimeseries"
              @finished-loading="onMapFinishedLoading"
            />
          </section>

          <footer
            v-if="!isBusy"
            class="d-flex align-items-end d-flex justify-content-between mt-1 mb-0"
          >
            <div v-if="!isGeoView" class="flex-grow-1">
              <data-size
                :current-size="numRows"
                :total="totalNumRows"
                @submit="onDataSizeSubmit"
              />
              <strong class="matching-color">matching</strong> samples of
              {{ totalNumRows }} to model<template v-if="selectionNumRows > 0">
                , {{ selectionNumRows }}
                <strong class="selected-color">selected</strong>
              </template>
            </div>
            <div v-else class="flex-grow-1">
              <p class="m-0">
                Selected Area Coverage:
                <strong class="matching-color">
                  {{ areaCoverage }}km<sup>2</sup>
                </strong>
              </p>
            </div>
            <b-button-toolbar v-if="isSelectState">
              <b-button-group class="ml-2 mt-1">
                <b-button
                  :variant="include ? 'primary' : 'secondary'"
                  @click="setIncludedActive"
                >
                  Included
                </b-button>
                <b-button
                  class="exclude-button"
                  :variant="!include ? 'primary' : 'secondary'"
                  @click="setExcludedActive"
                >
                  Excluded
                </b-button>
              </b-button-group>
            </b-button-toolbar>
            <!-- RESULT AND PREDICTION VIEW COMPONENTS-->

            <predictions-data-uploader
              :fitted-solution-id="fittedSolutionId"
              :target="targetName"
              :target-type="targetType"
              @model-apply="onApplyModel"
            />
            <save-modal
              ref="saveModel"
              :solution-id="solutionId"
              :fitted-solution-id="fittedSolutionId"
              @save="onSaveModel"
            />
            <forecast-horizon
              v-if="isTimeseries"
              :dataset="dataset"
              :fitted-solution-id="fittedSolutionId"
              :target="targetName"
              :target-type="targetType"
              @model-apply="onApplyModel"
            />
          </footer>
        </template>
      </main>
      <left-side-panel
        v-if="isOutcomeToggled"
        panel-title="Outcome Variables"
        class="overflow-auto"
      >
        <div v-if="state.name === 'result'">
          <error-threshold-slider v-if="showResiduals && !isTimeseries" />
          <result-facets
            :single-solution="isSingleSolution"
            :show-residuals="showResiduals"
            @fetch-summary-solution="fetchSummarySolution"
          />
        </div>
        <facet-list-pane
          v-else-if="isLabelState"
          :variables="secondaryVariables"
          :enable-color-scales="geoVarExists"
          :include="include"
          :summaries="secondarySummaries"
          :enable-footer="isSelectState"
          :dataset="dataset"
          instance-name="outcome-variables"
          @fetch-summaries="fetchSummaries"
        />
        <prediction-summaries
          v-else
          :is-busy="dataLoading"
          @fetch-summary-prediction="fetchSummaryPrediction"
        />
      </left-side-panel>
      <status-sidebar />
      <status-panel :dataset="dataset" />
      <b-modal :id="labelModalId" @ok="onLabelSubmit">
        <template #modal-header>
          {{ labelModalTitle }}
        </template>
        <b-form-group
          v-if="!hasLabelRole"
          id="input-group-1"
          label="Label name:"
          label-for="label-input-field"
          description="Enter the name of label."
          invalid-feedback="Label Name is Required"
        >
          <b-form-input
            id="label-input-field"
            v-model="labelName"
            type="text"
            required
            :placeholder="labelName"
            :state="labelNameState"
          />
        </b-form-group>
        <b-form-group
          v-else
          label="Label name:"
          label-for="label-select-field"
          description="Select the label field."
        >
          <b-form-select
            id="label-select-field"
            v-model="labelName"
            :options="options"
          />
        </b-form-group>
      </b-modal>
      <b-modal
        :id="unsaveModalId"
        ok-variant="danger"
        ok-title="Delete Cloned Dataset"
        @ok="onConfirmRouteSave(nextRoute)"
        @cancel="onCancelRouteSave(nextRoute)"
      >
        <template #modal-header> Unsaved dataset </template>
        <template>
          Current dataset is unsaved, are you sure you want to continue?
        </template>
      </b-modal>
      <save-dataset
        modal-id="save-dataset-modal"
        :dataset-name="dataset"
        :summaries="summaries"
      />
    </div>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import { capitalize } from "lodash";

// Components
import ActionColumn from "../components/layout/ActionColumn.vue";
import AddVariablePane from "../components/panel/AddVariablePane.vue";
import ExportPane from "../components/panel/ExportPane.vue";
import CreateLabelingForm from "../components/labelingComponents/CreateLabelingForm.vue";
import DataSize from "../components/buttons/DataSize.vue";
import ErrorThresholdSlider from "../components/ErrorThresholdSlider.vue";
import FacetListPane from "../components/panel/FacetListPane.vue";
import ForecastHorizon from "../components/ForecastHorizon.vue";
import GeoPlot from "../components/GeoPlot.vue";
import ImageMosaic from "../components/ImageMosaic.vue";
import LabelHeaderButtons from "../components/labelingComponents/LabelHeaderButtons.vue";
import ColorScaleSelection from "../components/ColorScaleSelection.vue";
import LayerSelection from "../components/LayerSelection.vue";
import LeftSidePanel from "../components/layout/LeftSidePanel.vue";
import LegendWeight from "../components/LegendWeight.vue";
import LoadingSpinner from "../components/LoadingSpinner.vue";
import PredictionsDataUploader from "../components/PredictionsDataUploader.vue";
import PredictionSummaries from "../components/PredictionSummaries.vue";
import ResultFacets from "../components/ResultFacets.vue";
import SaveDataset from "../components/labelingComponents/SaveDataset.vue";
import SaveModal from "../components/SaveModal.vue";
import SearchBar from "../components/layout/SearchBar.vue";
import SelectDataTable from "../components/SelectDataTable.vue";
import SelectGraphView from "../components/SelectGraphView.vue";
import SelectTimeseriesView from "../components/SelectTimeseriesView.vue";
import StatusPanel from "../components/StatusPanel.vue";
import StatusSidebar from "../components/StatusSidebar.vue";
// Store
import { viewActions, datasetActions, datasetGetters } from "../store";
import { Variable } from "../store/dataset/index";
import { DATA_EXPLORER_VAR_INSTANCE } from "../store/route/index";
import { getters as routeGetters } from "../store/route/module";

// Util
import { overlayRouteEntry } from "../util/routes";
import { META_TYPES } from "../util/types";
import { SelectViewState } from "../util/state/AppStateWrapper";
import {
  bindMethods,
  SelectViewConfig,
  genericMethods,
  genericComputes,
  labelMethods,
  labelComputes,
  resultMethods,
  resultComputes,
  selectComputes,
  selectMethods,
  predictionMethods,
  predictionComputes,
  ExplorerViewComponent,
} from "../util/explorer";
import { findAPositiveLabel } from "../util/data";
import _ from "lodash";
import { DataExplorerRef } from "../util/componentTypes";
import { GENERIC_EVENT_HANDLERS } from "../util/explorer/functions/generic";

const DataExplorer = Vue.extend({
  name: "DataExplorer",

  components: {
    ActionColumn,
    AddVariablePane,
    ExportPane,
    CreateLabelingForm,
    DataSize,
    ErrorThresholdSlider,
    FacetListPane,
    ForecastHorizon,
    GeoPlot,
    ImageMosaic,
    LabelHeaderButtons,
    ColorScaleSelection,
    LayerSelection,
    LeftSidePanel,
    LegendWeight,
    LoadingSpinner,
    PredictionsDataUploader,
    PredictionSummaries,
    ResultFacets,
    SaveDataset,
    SaveModal,
    SearchBar,
    SelectDataTable,
    SelectGraphView,
    SelectTimeseriesView,
    StatusPanel,
    StatusSidebar,
  },

  data() {
    return {
      activeView: ExplorerViewComponent.TABLE, // TABLE_VIEW
      busyState: "Busy", // contains the info to display to the user when the UI is busy
      config: new SelectViewConfig(), // this config controls what is displayed in the action bar
      dataLoading: false, // this controls the spinners for the data view tabs (table, mosaic, geoplot)
      include: true, // this controls the include exclude view for the select state
      instanceName: DATA_EXPLORER_VAR_INSTANCE, // component instance name
      isBusy: false, // controls spinners in label state when search similar or save is used
      labelModalId: "label-input-form", // modal id
      shouldSaveDataset: false, // There is not a not a nice way of denoting saving currently for the label view
      unsaveModalId: "unsaved-modal",
      labelName: "", // labelName of the variable being annotated in the label view
      labelNameState: null,
      metaTypes: Object.keys(META_TYPES), // all of the meta types categories
      state: new SelectViewState(), // this state controls data flow,
      nextRoute: null,
      observer: null,
    };
  },

  // Update either the summaries or explore data on user interaction.
  watch: {
    async solutionId() {
      this.dataLoading = true;
      await this.state.fetchData();
      this.dataLoading = false;
    },

    async produceRequestId() {
      this.isBusy = true;
      this.dataLoading = true;
      this.activeView = ExplorerViewComponent.TABLE;
      this.busyState = "Fetching Variables";
      await this.state.fetchVariables();
      this.busyState = "Fetch Summaries";
      await this.state.fetchVariableSummaries();
      this.busyState = "Fetching Data";
      await this.state.fetchData();
      await this.state.fetchMapBaseline();
      this.dataLoading = false;
      this.isBusy = false;
      this.busyState = "Busy";
    },

    async activeVariables(n, o) {
      if (_.isEqual(n, o)) return;
      await this.state.fetchVariableSummaries();
    },

    async filters(n, o) {
      if (n === o) return;
      this.dataLoading = true;
      await this.state.fetchData();
      this.dataLoading = false;
    },

    async highlights(n, o) {
      if (_.isEqual(n, o)) return;
      this.dataLoading = true;
      await this.state.fetchData();
      this.dataLoading = false;
    },

    async explore(n, o) {
      if (_.isEqual(n, o)) return;
      this.dataLoading = true;
      await viewActions.updateDataExplorerData(this.$store);
      this.dataLoading = false;
    },

    async geoVarExists() {
      const self = (this as unknown) as DataExplorerRef; // because the computes/methods are added in beforeCreate typescript does not work so we cast it to a type here
      if (
        (!self.geoVarExists && self.summaries.some((s) => s.pending)) ||
        self.geoVarExists === routeGetters.hasGeoData(this.$store)
      ) {
        return;
      }
      const route = routeGetters.getRoute(this.$store);
      const entry = overlayRouteEntry(route, { hasGeoData: self.geoVarExists });
      this.$router.push(entry).catch((err) => console.warn(err));
    },

    targetName() {
      const self = (this as unknown) as DataExplorerRef; // because the computes/methods are added in beforeCreate typescript does not work so we cast it to a type here
      datasetActions.fetchOutliers(this.$store, self.dataset);
      // if binary classification add positive label
      if (routeGetters.isBinaryClassification(this.$store)) {
        //find target summary
        const targetSummary = self.summaries.find(
          (v) => v.key === self.target.key
        );
        if (targetSummary) {
          // build labels from buckets
          const buckets = targetSummary?.baseline?.buckets;
          if (buckets) {
            // use the buckets keys as labels
            const labels = buckets.map((bucket) => bucket.key);
            if (labels.length === 2) {
              // get the "most positive" label
              const positiveLabel = findAPositiveLabel(labels);
              self.updateRoute({ positiveLabel });
            }
          }
        }
      }
      // fetch metrics
      const metrics = routeGetters.getModelMetrics(this.$store);
      if (metrics) {
        const storedMetrics = datasetGetters.getModelingMetrics(this.$store);
        if (!storedMetrics.some((m) => m.displayName === metrics[0])) {
          self.updateRoute({ metrics: "" });
        }
      }
    },
  },

  async beforeRouteLeave(to, from, next) {
    // react to route changes...
    // don't forget to call next()
    const self = (this as unknown) as DataExplorerRef;
    this.nextRoute = next;

    if (self.isClone) {
      const isDatasetSaved = await self.isCurrentDatasetSaved();

      if (!isDatasetSaved) {
        // show dialog
        self.$bvModal.show(this.unsaveModalId);
      } else {
        next();
      }
    } else {
      next();
    }
  },
  beforeCreate() {
    const self = (this as unknown) as DataExplorerRef; // because the computes/methods are added in beforeCreate typescript does not work so we cast it to a type here
    // computes / methods need to be binded to the instance
    this.$options.computed = {
      ...this.$options.computed, // any computes defined in the component
      ...bindMethods(genericComputes, self), // generic computes used across all states
      ...bindMethods(resultComputes, self), // computes used in result state
      ...bindMethods(selectComputes, self), // computes used in select state
      ...bindMethods(predictionComputes, self), // computes used in prediction state
      ...bindMethods(labelComputes, self), // computes used in the label state
    };
    // methods for each state need to be bound to the DataExplorer instance
    this.$options.methods = {
      ...this.$options.methods, // any methods defined in the component
      ...bindMethods(genericMethods, self), // generic computes used across all states
      ...bindMethods(selectMethods, self), // computes used in result state
      ...bindMethods(labelMethods, self), // computes used in select state
      ...bindMethods(resultMethods, self), // computes used in prediction state
      ...bindMethods(predictionMethods, self), // computes used in the label state
    };
  },
  beforeDestroy() {
    this.removeEventHandlers(this.config.eventHandlers);
    this.removeEventHandlers(GENERIC_EVENT_HANDLERS);
  },
  async beforeMount() {
    const self = (this as unknown) as DataExplorerRef; // because the computes/methods are added in beforeCreate typescript does not work so we cast it to a type here
    self.bindEventHandlers(GENERIC_EVENT_HANDLERS);
    if (self.isSelectState) {
      // First get the dataset informations
      await viewActions.fetchDataExplorerData(this.$store, [] as Variable[]);
      // Pre-select the top 5 variables by importance
      self.preSelectTopVariables();
      // Update the explore data
      await viewActions.updateDataExplorerData(this.$store);
    }
  },

  mounted() {
    const self = (this as unknown) as DataExplorerRef; // because the computes/methods are added in beforeCreate typescript does not work so we cast it to a type here
    self.changeStatesByName(self.explorerRouteState);
    self.labelName = routeGetters.getRouteLabel(this.$store);
  },

  methods: {
    capitalize,
    bindEventHandlers(eventHandlers: Record<string, Function>) {
      const self = (this as unknown) as DataExplorerRef; // because the computes/methods are added in beforeCreate typescript does not work so we cast it to a type here
      // get event names the functions are listening for
      const eventKeys = Object.keys(eventHandlers);
      // bind the function to the instance of this component
      const boundedEventHandlers = bindMethods(eventHandlers, self);
      // apply them to the global event bus
      eventKeys.forEach((event) => {
        this.$eventBus.$on(event, boundedEventHandlers[event]);
      });
      return;
    },
    removeEventHandlers(eventHandlers: Record<string, Function>) {
      // get list of events being listened to
      const eventKeys = Object.keys(eventHandlers);
      // remove all listeners for these events
      eventKeys.forEach((event) => {
        this.$eventBus.$off(event);
      });
    },
  },
});
export default DataExplorer;
</script>

<style scoped>
.view-container {
  display: flex;
  flex-direction: row;
  flex-wrap: nowrap;
  flex-grow: 1;
  overflow: hidden;
  height: calc(var(--content-full-height)-56px);
}

.nav-bar-margin {
  margin-top: var(--navbar-outer-height);
  height: var(--content-full-height);
}

/* Make some elements of a container unsquishable. */
.view-container > *:not(.content),
.content > *:not(.data-container) {
  flex-shrink: 0;
}

.content {
  display: flex;
  flex-direction: column;
  flex-grow: 1;
  padding-bottom: 1rem;
  padding-top: 1rem;
}

/* Add padding to all elements but the tabs and data */
.content > *:not(.data-container),
.content > *:not(.tab-container) {
  padding-left: 1rem;
  padding-right: 1rem;
}

.tab-container,
.data-container {
  border-bottom: 1px solid var(--border-color);
}

.data-container {
  background-color: var(--white);
  display: flex;
  flex-flow: wrap;
  height: 100%;
  padding: 1rem;
  position: relative;
  width: 100%;
}
</style>
<style>
.view-container .tab-container ul.nav-tabs {
  border: none;
  margin-bottom: -1px;
}

.view-container .tab-container a.nav-link {
  border: 1px solid transparent;
  border-bottom-color: var(--border-color);
  border-top-width: 3px;
  color: var(--color-text-second);
  margin-bottom: 0;
}

.view-container .tab-container a.nav-link.active {
  background-color: var(--white);
  border-color: var(--border-color);
  border-top-color: var(--primary);
  border-bottom-width: 0;
  border-top-left-radius: 0.25rem;
  border-top-right-radius: 0.25rem;
  color: var(--primary);
  margin-bottom: -1px;
}

.select-data-action-exclude:not([disabled]) .include-highlight,
.select-data-action-exclude:not([disabled]) .exclude-highlight {
  color: var(--blue); /* #255dcc; */
}

.select-data-action-exclude:not([disabled]) .include-selection,
.select-data-action-exclude:not([disabled]) .exclude-selection {
  color: var(--red); /* #ff0067; */
}
</style>
