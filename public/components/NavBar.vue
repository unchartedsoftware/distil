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
  <b-navbar toggleable="lg" type="dark" class="fixed-top">
    <b-nav-toggle target="nav-collapse" />

    <!-- Branding -->
    <img
      src="/images/uncharted.svg"
      class="app-icon"
      height="36"
      width="36"
      :title="version"
    />
    <b-navbar-brand>Distil</b-navbar-brand>

    <!-- Left Side -->
    <b-collapse id="nav-collapse" is-nav>
      <b-navbar-nav>
        <b-nav-item :active="isActive(SEARCH_ROUTE)" @click="onSearch">
          <i class="fa fa-home nav-icon" /> Select Model or Dataset
        </b-nav-item>

        <!-- If search produces a model of interest, select it for reuse: will start Apply Model workflow. -->
        <template v-if="isApplyModel && !isActive(DATA_EXPLORER_ROUTE)">
          <b-nav-item
            :active="isActive(APPLY_MODEL_ROUTE)"
            @click="onApplyModel"
          >
            <i class="fa fa-table nav-icon" /> Apply Model: Select New Data
          </b-nav-item>

          <b-nav-item
            :active="isActive(PREDICTION_ROUTE)"
            :disabled="isActive(APPLY_MODEL_ROUTE)"
            @click="onPredictions"
          >
            <i class="fa fa-line-chart nav-icon" /> View Predictions
          </b-nav-item>
        </template>

        <!-- If no appropriate model exist, select a dataset: will start New Model workflow. -->
        <template v-else-if="isActive(DATA_EXPLORER_ROUTE)">
          <b-nav-item
            :active="explorerSelectState && target === null"
            @click="onSelectTarget()"
          >
            <i class="fa fa-crosshairs nav-icon" /> Select Target
          </b-nav-item>
          <b-nav-item
            v-if="target !== null"
            :active="explorerSelectState && target !== null"
            class="d-flex"
            @click="explorerNav('select')"
          >
            <i
              class="fas fa-dumbbell fa-sm nav-icon d-inline-flex align-items-center justify-content-center"
            />
            Select Training
          </b-nav-item>
          <b-nav-item
            v-if="haveSolutions"
            :active="explorerResultState"
            @click="explorerNav('result')"
          >
            <i class="fa fa-check-circle nav-icon" /> Check Models
          </b-nav-item>
          <b-nav-item
            v-if="havePredictions"
            :active="explorerPredictionState"
            @click="explorerNav('prediction')"
          >
            <i class="fa fa-line-chart nav-icon" /> View Predictions
          </b-nav-item>
          <b-nav-item v-if="explorerLabelState" :active="explorerLabelState">
            <i class="fa fa-tag nav-icon" /> Label Data
          </b-nav-item>
        </template>
      </b-navbar-nav>
    </b-collapse>

    <!--<b-nav-item
      @click="onJoinDatasets"
      v-if="isJoinDatasets && isActive(JOIN_DATASETS_ROUTE)"
      :active="isActive(JOIN_DATASETS_ROUTE)"
    >
      <i class="fa fa-database nav-icon"></i> Join Datasets
    </b-nav-item>-->
    <div>
      <!-- Disabled buttons are not emitting mouse over events so we use an outer div -->
      <!-- select target hint -->
      <div
        v-if="!isActive(SEARCH_ROUTE) && explorerSelectState && target === null"
        @mouseenter="onHoverHint(EventList.HINTS.SELECT_TARGET)"
        @mouseleave="cancelHint(EventList.CANCEL_HINTS.SELECT_TARGET)"
      >
        <b-button variant="primary" disabled> Select Target </b-button>
      </div>
      <div
        v-if="
          explorerSelectState && target !== null && !trainingVariables.length
        "
        @mouseenter="onHoverHint(EventList.HINTS.SELECT_TRAINING)"
        @mouseleave="cancelHint(EventList.HINTS.SELECT_TRAINING)"
      >
        <b-button variant="primary" disabled> Select Training </b-button>
      </div>

      <create-solutions-form
        v-if="
          explorerSelectState && target !== null && trainingVariables.length > 0
        "
        :aria-disabled="isCreateModelPossible"
        class="ml-2"
      />
      <!--PREDICTION STATE SECTION-->
      <b-button v-if="explorerPredictionState" v-b-modal.save class="mr-1">
        Create Dataset
      </b-button>
      <b-button
        v-if="explorerPredictionState"
        v-b-modal.export
        variant="primary"
      >
        Export Predictions
      </b-button>
      <!--RESULT STATE SECTION-->
      <template
        v-if="
          explorerResultState && (isSingleSolution || isActiveSolutionSaved)
        "
      >
        <b-button
          v-if="isTimeseries"
          variant="success"
          class="apply-button"
          @click="$bvModal.show('forecast-horizon-modal')"
        >
          Forecast
        </b-button>
        <b-button
          v-else
          variant="success"
          class="apply-button"
          @click="$bvModal.show('predictions-data-upload-modal')"
        >
          Apply Model
        </b-button>
      </template>
      <b-button
        v-else-if="explorerResultState"
        v-b-modal.save-model-modal
        :disabled="!currentSolutionCompleted"
        variant="success"
        class="save-button"
      >
        <i class="fa fa-floppy-o" />
        Save Model
      </b-button>
      <create-labeling-form
        v-if="explorerLabelState"
        class="d-flex justify-content-between h-100 align-items-center"
        :low-shot-summary="labelSummary"
      />
    </div>
    <!-- Right side -->
    <b-navbar-nav class="ml-auto">
      <b-nav-item :href="helpURL">Help</b-nav-item>
    </b-navbar-nav>
  </b-navbar>
</template>

<script lang="ts">
import "../assets/images/uncharted.svg";
import {
  gotoApplyModel,
  // gotoHome,
  gotoJoinDatasets,
  gotoPredictions,
  gotoResults,
  gotoSearch,
  gotoSelectData,
} from "../util/nav";
import { appGetters, datasetGetters, requestGetters } from "../store";
import { getters as routeGetters } from "../store/route/module";
import {
  APPLY_MODEL_ROUTE,
  // HOME_ROUTE,
  SEARCH_ROUTE,
  JOIN_DATASETS_ROUTE,
  SELECT_TARGET_ROUTE,
  SELECT_TRAINING_ROUTE,
  RESULTS_ROUTE,
  PREDICTION_ROUTE,
  DATA_EXPLORER_ROUTE,
} from "../store/route/index";
import { isFittedSolutionIdSavedAsModel } from "../util/models";
import { restoreView } from "../util/view";
import Vue from "vue";
import { ExplorerStateNames } from "../util/explorer";
import { EventList } from "../util/events";
import { createRouteEntry } from "../util/routes";
import { Variable, VariableSummary } from "../store/dataset";
import { getSolutionById } from "../util/solutions";
import { Solution, SolutionStatus } from "../store/requests";
// components
import CreateSolutionsForm from "../components/CreateSolutionsForm.vue";
import CreateLabelingForm from "../components/labelingComponents/CreateLabelingForm.vue";
import {
  getAllVariablesSummaries,
  hasRole,
  LOW_SHOT_RANK_COLUMN_PREFIX,
  LOW_SHOT_SCORE_COLUMN_PREFIX,
} from "../util/data";
import { DISTIL_ROLES } from "../util/types";

export default Vue.extend({
  name: "NavBar",
  components: {
    CreateSolutionsForm,
    CreateLabelingForm,
  },
  data() {
    return {
      APPLY_MODEL_ROUTE: APPLY_MODEL_ROUTE,
      // HOME_ROUTE: HOME_ROUTE,
      SEARCH_ROUTE: SEARCH_ROUTE,
      JOIN_DATASETS_ROUTE: JOIN_DATASETS_ROUTE,
      SELECT_TARGET_ROUTE: SELECT_TARGET_ROUTE,
      SELECT_TRAINING_ROUTE: SELECT_TRAINING_ROUTE,
      RESULTS_ROUTE: RESULTS_ROUTE,
      PREDICTION_ROUTE: PREDICTION_ROUTE,
      DATA_EXPLORER_ROUTE: DATA_EXPLORER_ROUTE,
    };
  },

  computed: {
    path(): string {
      return routeGetters.getRoutePath(this.$store);
    },

    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },

    target(): string {
      return routeGetters.getRouteTargetVariable(this.$store);
    },

    joinDatasets(): string[] {
      return routeGetters.getRouteJoinDatasets(this.$store);
    },

    joinDatasetsHash(): string {
      return routeGetters.getRouteJoinDatasetsHash(this.$store);
    },

    isJoinDatasets(): boolean {
      return this.joinDatasets.length === 2 || this.hasJoinDatasetView();
    },
    dataExplorerState(): ExplorerStateNames {
      return routeGetters.getDataExplorerState(this.$store);
    },
    isApplyModel(): boolean {
      /*
        Check if we requested in the route for an Apply Model navigation,
        or, in the case of a prediction a fitted solution ID.
       */
      return (
        routeGetters.isApplyModel(this.$store) ||
        !!routeGetters.getRouteFittedSolutionId(this.$store)
      );
    },
    explorerSelectState(): boolean {
      return this.dataExplorerState === ExplorerStateNames.SELECT_VIEW;
    },
    explorerResultState(): boolean {
      return this.dataExplorerState === ExplorerStateNames.RESULT_VIEW;
    },
    explorerPredictionState(): boolean {
      return this.dataExplorerState === ExplorerStateNames.PREDICTION_VIEW;
    },
    explorerLabelState(): boolean {
      return this.dataExplorerState === ExplorerStateNames.LABEL_VIEW;
    },
    trainingVariables(): Variable[] {
      return routeGetters.getTrainingVariables(this.$store) ?? [];
    },
    hasDataset(): boolean {
      return !!this.dataset;
    },
    helpURL(): string {
      return appGetters.getHelpURL(this.$store);
    },
    haveSolutions(): boolean {
      return requestGetters.getRelevantSolutions(this.$store).length > 0;
    },
    havePredictions(): boolean {
      return requestGetters.getRelevantSolutions(this.$store).some((p) => {
        return p.hasPredictions;
      });
    },
    isActiveSolutionSaved(): boolean | undefined {
      return isFittedSolutionIdSavedAsModel(
        requestGetters.getActiveSolution(this.$store)?.fittedSolutionId
      );
    },
    isSingleSolution(): boolean {
      return routeGetters.isSingleSolution(this.$store);
    },
    version(): string {
      return appGetters.getAllSystemVersions(this.$store);
    },
    isTimeseries(): boolean {
      return routeGetters.isTimeseries(this.$store);
    },
    currentSolutionCompleted(): boolean {
      let solutionRequests = requestGetters.getRelevantSolutionRequests(
        this.$store
      );
      let solutions = [] as Solution[];
      const solutionId = routeGetters.getRouteSolutionId(this.$store);
      if (this.isSingleSolution) {
        const solution = getSolutionById(
          requestGetters.getSolutions(this.$store),
          solutionId
        );
        if (solution) {
          solutions = [solution];
          solutionRequests = [
            solutionRequests.find(
              (request) => request.requestId === solution.requestId
            ),
          ];
        }
      } else {
        // multiple solutions
        solutions = requestGetters.getRelevantSolutions(this.$store);
      }

      return solutions.some(
        (s) =>
          s.solutionId === solutionId &&
          s.progress === SolutionStatus.SOLUTION_COMPLETED
      );
    },
    variables(): Variable[] {
      const labelName = routeGetters.getRouteLabel(this.$store);
      const labelScoreName = LOW_SHOT_SCORE_COLUMN_PREFIX + labelName;
      const labelRankName = LOW_SHOT_RANK_COLUMN_PREFIX + labelName;
      return datasetGetters.getVariables(this.$store).filter((v) => {
        return (
          v.key !== labelScoreName &&
          v.key !== labelRankName &&
          !hasRole(v, DISTIL_ROLES.SystemData)
        );
      });
    },
    summaries(): VariableSummary[] {
      const summaryDictionary = datasetGetters.getVariableSummariesDictionary(
        this.$store
      );
      const dataset = routeGetters.getRouteDataset(this.$store);
      return getAllVariablesSummaries(
        this.variables,
        summaryDictionary,
        dataset
      );
    },
    labelSummary(): VariableSummary {
      const label = routeGetters.getRouteLabel(this.$store);
      return this.summaries.find((s) => {
        return s.key === label;
      });
    },
    isCreateModelPossible(): boolean {
      const training = routeGetters.getTrainingVariables(this.$store);
      // check that we have some target and training variables.
      return this.target != null && training?.length > 0;
    },
  },

  methods: {
    explorerNav(state: string) {
      this.$emit(EventList.EXPLORER.NAV_EVENT, state);
    },
    onSelectTarget() {
      const routeDataset = routeGetters.getRouteDataset(this.$store);
      const exploreVariables = routeGetters.getExploreVariables(this.$store);
      const entry = createRouteEntry(DATA_EXPLORER_ROUTE, {
        dataset: routeDataset,
        explore: exploreVariables.join(","),
      });
      this.$router.push(entry).catch((err) => console.warn(err));
      this.explorerNav(ExplorerStateNames.SELECT_VIEW);
    },
    isActive(view) {
      return view === this.path;
    },
    isState(state: ExplorerStateNames): boolean {
      return state === this.dataExplorerState;
    },
    onHoverHint(hint: EventList.HINTS) {
      this.$eventBus.$emit(hint, true);
    },
    cancelHint(hint: EventList.HINTS) {
      this.$eventBus.$emit(hint, false);
    },
    // onHome() {
    //   gotoHome(this.$router);
    // },

    onSearch() {
      gotoSearch(this.$router);
    },

    onJoinDatasets() {
      gotoJoinDatasets(this.$router);
    },

    onSelectData() {
      gotoSelectData(this.$router);
    },

    onResults() {
      gotoResults(this.$router);
    },

    onApplyModel() {
      gotoApplyModel(this.$router);
    },

    onPredictions() {
      gotoPredictions(this.$router);
    },

    hasJoinDatasetView(): boolean {
      return !!restoreView(JOIN_DATASETS_ROUTE, this.joinDatasetsHash);
    },
  },
});
</script>

<style scoped>
.navbar {
  background-color: var(--gray-900);
  box-shadow: 0 6px 12px 0 rgba(0, 0, 0, 0.1);
  justify-content: flex-start;
}

.app-icon {
  margin-right: 0.33em;
}

.app-icon.is-prototype {
  filter: invert(1);
}

.nav-item {
  font-weight: bold;
  letter-spacing: 0.01rem;
  white-space: nowrap;
}

/* Display an arrow if two link are next to each others. */
.navbar-collapse:not(.show) .nav-item + .nav-item .nav-link::before,
.navbar-collapse.show .nav-item + .nav-item::before {
  color: var(--gray-600);
  font-family: FontAwesome;
  font-weight: bold;
}

/* Horizontal arrow if the menu is visible (not collapsed). */
.navbar-collapse:not(.show) .nav-item + .nav-item .nav-link::before {
  font-family: "Font Awesome\ 5 Free";
  content: "\f715"; /* angle-right => https://fontawesome.com/v4.7.0/cheatsheet/ */
  margin-right: 1em;
  font-weight: 900;
  transform: rotate(90deg);
  display: inline-block;
}

/* Change the arrow to be vertical if the menu is collapsed. */
.navbar-collapse.show .nav-item + .nav-item {
  position: relative;
  margin-top: 1em;
}
.navbar-collapse.show .nav-item + .nav-item::before {
  content: "\f107"; /* angle-down => https://fontawesome.com/v4.7.0/cheatsheet/ */
  left: 0.65em;
  position: absolute;
  top: -1em;
}

/* Icon. */
.nav-icon {
  border-radius: 50%;
  height: 30px;
  margin-right: 0.25em;
  padding: 7px;
  text-align: center;
  width: 30px;
}

/*
  In the following I use the ID #distil-app to overwrite the Bootstrap CSS
  by increasing the selectors specificity.
*/

/* Default colours */
#distil-app .nav-link {
  transition: color 0.25s;
  color: var(--gray-600);
}
#distil-app .nav-link .nav-icon {
  transition: background 0.25s, color 0.25s;
  background-color: var(--gray-800);
  color: var(--gray-400);
}

/* Active and non disabled on hover nav-item */
#distil-app .nav-link.active,
#distil-app .nav-link:not(.disable):hover {
  color: var(--gray-400);
}
#distil-app .nav-link.active .nav-icon,
#distil-app .nav-link:not(.disable):hover .nav-icon {
  background-color: var(--black);
}

/* Disabled Nav-item */
#distil-app .nav-link.disabled .nav-icon {
  background: none;
  color: var(--gray-600);
}
</style>
