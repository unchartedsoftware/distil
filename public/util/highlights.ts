import { Highlight, HighlightRoot } from '../store/highlights/index';
import { Filter, NUMERICAL_FILTER } from '../util/filters';
import { getters as routeGetters } from '../store/route/module';
import { getters as highlightGetters } from '../store/highlights/module';
import { overlayRouteEntry } from '../util/routes'
import { FilterParams } from '../util/filters'
import { getFilterType, getVarType, isFeatureType, addFeaturePrefix, isClusterType, addClusterPrefix } from '../util/types'
import _ from 'lodash';
import { store } from '../store/storeProvider';
import VueRouter from 'vue-router';

export function encodeHighlights(highlightRoot: HighlightRoot): string {
	if (_.isEmpty(highlightRoot)) {
		return null;
	}
	return btoa(JSON.stringify(highlightRoot));
}

export function decodeHighlights(highlightRoot: string): HighlightRoot {
	if (_.isEmpty(highlightRoot)) {
		return null;
	}
	return JSON.parse(atob(highlightRoot)) as HighlightRoot;
}

export function createFilterFromHighlightRoot(highlightRoot: HighlightRoot, mode: string): Filter {
	if (!highlightRoot || highlightRoot.value == null) {
		return null;
	}
	// inject metadata prefix for metadata vars
	let key = highlightRoot.key;
	const type = getVarType(key);
	if (isFeatureType(type)) {
		key = addFeaturePrefix(key);
	}
	if (isClusterType(type)) {
		key = addClusterPrefix(key);
	}
	const filterType = getFilterType(type);
	if (_.isString(highlightRoot.value)) {
		return {
			key: key,
			type: filterType,
			mode: mode,
			categories: [highlightRoot.value]
		};
	}
	if (highlightRoot.value.from !== undefined && highlightRoot.value.to !== undefined) {
		return {
			key: key,
			type: NUMERICAL_FILTER,
			mode: mode,
			min: highlightRoot.value.from,
			max: highlightRoot.value.to
		};
	}
	return null;
}

export function addHighlightToFilterParams(filterParams: FilterParams, highlightRoot: HighlightRoot, mode: string): FilterParams {
	const params = _.cloneDeep(filterParams);
	const highlightFilter = createFilterFromHighlightRoot(highlightRoot, mode);
	if (highlightFilter) {
		params.filters.push(highlightFilter);
	}
	return params;
}

export function updateHighlightRoot(router: VueRouter, highlightRoot: HighlightRoot) {
	const entry = overlayRouteEntry(routeGetters.getRoute(store()), {
		highlights: encodeHighlights(highlightRoot),
		row: null // clear row
	});
	router.push(entry);
}

export function clearHighlightRoot(router: VueRouter) {
	const entry = overlayRouteEntry(routeGetters.getRoute(store()), {
		highlights: null,
		row: null // clear row
	});
	router.push(entry);
}

export function getHighlights(): Highlight {
	return {
		root: routeGetters.getDecodedHighlightRoot(store()),
		values: {
			summaries: highlightGetters.getHighlightedSummaries(store())
		}
	};
}
