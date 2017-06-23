import _ from 'lodash';

export function setDatasets(state, datasets) {
	state.datasets = datasets;
}

export function addDataset(state, dataset) {
	state.datasets.push(dataset);
}

export function removeDataset(state, id) {
	return !_.isUndefined(_.remove(state.datasets, elem => elem.name === id));
}

export function setActiveDataset(state, id) {
	state.activeDataset = id;
}

export function setVariableSummaries(state, summaries) {
	state.variableSummaries = summaries;
}

export function updateVariableSummaries(state, args) {
	state.variableSummaries.splice(args.index, 1);
	state.variableSummaries.splice(args.index, 0, args.histogram);
}

export function setFilteredData(state, filteredData) {
	state.filteredData = filteredData;
}
