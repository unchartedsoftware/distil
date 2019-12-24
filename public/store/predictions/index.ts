import { Dictionary } from "../../util/dict";
import { VariableSummary, Extrema, TableData } from "../dataset/index";

export interface PredictionState {
  // table data
  includedResultTableData: TableData;
  excludedResultTableData: TableData;
  // training / target
  trainingSummaries: VariableSummary[];
  targetSummary: VariableSummary;
  // predicted
  predictedSummaries: VariableSummary[];
  // residuals
  residualSummaries: VariableSummary[];
  residualsExtrema: Extrema;
  // correctness summary (correct vs. incorrect) for predicted categorical data
  correctnessSummaries: VariableSummary[];
  // forecasts
  timeseries: Dictionary<Dictionary<number[][]>>;
  forecasts: Dictionary<Dictionary<number[][]>>;
  fittedSolutionId: string;
  produceRequestId: string;
}

export const state: PredictionState = {
  // table data
  includedResultTableData: null,
  excludedResultTableData: null,
  // training / target
  trainingSummaries: [],
  targetSummary: null,
  // predicted
  predictedSummaries: [],
  // residuals
  residualSummaries: [],
  residualsExtrema: { min: null, max: null },
  // correctness summary (correct vs. incorrect) for predicted categorical data
  correctnessSummaries: [],
  // forecasts
  timeseries: {},
  forecasts: {},
  fittedSolutionId: null,
  produceRequestId: null
};
