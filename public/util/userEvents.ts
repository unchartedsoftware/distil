// Activity, sub-activity and feature ID types used in the user event reporting logs

export enum Activity {
  DATA_PREPARATION = "DATA_PREPARATION",
  PROBLEM_DEFINITION = "PROBLEM_DEFINITION",
  MODEL_SELECTION = "MODEL_SELECTION"
}

export enum SubActivity {
  APP_LAUNCH = "APP_LAUNCH",
  DATA_OPEN = "DATA_OPEN",
  DATA_EXPLORATION = "DATA_EXPLORATION",
  DATA_AUGMENTATION = "DATA_AUGMENTATION",
  DATA_TRANSFORMATION = "DATA_TRANSFORMATION",
  PROBLEM_SPECIFICATION = "PROBLEM_SPECIFICATION",
  MODEL_SEARCH = "MODEL_SEARCH",
  MODEL_SUMMARIZATION = "MODEL_SUMMARIZATION",
  MODEL_COMPARISON = "MODEL_COMPARISON",
  MODEL_EXPLANATION = "MODEL_EXPLANATION",
  MODEL_EXPORT = "MODEL_EXPORT"
}

export enum Feature {
  SEARCH_DATASETS = "SEARCH_DATASETS",
  SELECT_DATASET = "SELECT_DATASET",
  SELECT_TARGET = "SELECT_TARGET",
  RETYPE_FEATURE = "RETYPE_FEATURE",
  RANK_FEATURES = "RANK_FEATURES",
  GEOCODE_FEATURES = "GEOCODE_FEATURES",
  JOIN_DATASETS = "JOIN_DATASETS",
  ADD_FEATURE = "ADD_FEATURE",
  ADD_ALL_FEATURES = "ADD_ALL_FEATURES",
  REMOVE_FEATURE = "REMOVE_FEATURE",
  REMOVE_ALL_FEATURES = "REMOVE_ALL_FEATURES",
  CHANGE_HIGHLIGHT = "CHANGE_HIGHLIGHT",
  CHANGE_SELECTION = "CHANGE_SELECTION",
  CHANGE_ERROR_THRESHOLD = "CHANGE_ERROR_THRESHOLD",
  FILTER_DATA = "FILTER_DATA",
  UNFILTER_DATA = "UNFILTER_DATA",
  SEARCH_FEATURES = "SEARCH_FEATURES",
  CREATE_MODEL = "CREATE_MODEL",
  SELECT_MODEL = "SELECT_MODEL",
  EXPORT_MODEL = "EXPORT_MODEL"
}
