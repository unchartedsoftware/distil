import _ from 'lodash';
import Vue from 'vue';
import { getters as routeGetters } from '../store/route/module';
import { overlayRouteEntry } from './routes';

/**
 * Categorical filter, omitting documents that do not contain the provided
 * categories in the variable.
 * @constant {string}
 */
export const CATEGORICAL_FILTER = 'categorical';

/**
 * Numerical filter, omitting documents that do not fall within the provided
 * variable range.
 * @constant {string}
 */
export const NUMERICAL_FILTER = 'numerical';

/**
 * Row filter, omitting documents that have the specified d3mIndices;
 * @constant {string}
 */
export const ROW_FILTER = 'row';

/**
 * Feature filter, omitting documents that have the specified feature value;
 * @constant {string}
 */
export const FEATURE_FILTER = 'feature';

/**
 * Include filter, excluding documents that do not fall within the filter.
 * @constant {string}
 */
export const INCLUDE_FILTER = 'include';

/**
 * Exclude filter, excluding documents that fall outside the filter.
 * @constant {string}
 */
export const EXCLUDE_FILTER = 'exclude';

export interface Filter {
	type: string;
	mode: string;
	name?: string;
	min?: number;
	max?: number;
	categories?: string[];
	d3mIndices?: string[];
}

export interface FilterParams {
	filters: Filter[];
	variables: string[];
	size?: number;
}

/**
 * Decodes the filters from the route string into an array.
 *
 * @param {string} filters - The filters from the route query string.
 *
 * @returns {Filter[]} The decoded filter object.
 */
export function decodeFilters(filters: string): Filter[] {
	if (_.isEmpty(filters)) {
		return [];
	}
	return JSON.parse(atob(filters)) as Filter[];
}

/**
 * Encodes the map of filter objects into a map of route query strings.
 *
 * @param {Filter[]} filters - The filter objects.
 *
 * @returns {string} The encoded route query strings.
 */
export function encodeFilters(filters: Filter[]): string {
	if (_.isEmpty(filters)) {
		return null;
	}
	return btoa(JSON.stringify(filters));
}

/**
 * Resolves any redundant row include / excludes such that there are only a
 * maximum of two row filters, one for includes, one for excludes.
 */
function dedupeRowFilters(filters: Filter[]): Filter[] {

	const selections = filters.filter(filter => filter.type === ROW_FILTER);
	const remaining = filters.filter(filter => filter.type !== ROW_FILTER);

	const included = {};
	const excluded = {};
	const d3mIndices = {};

	selections.forEach((filter, filterIndex) => {
		filter.d3mIndices.forEach(d3mIndex => {
			if (filter.mode === INCLUDE_FILTER) {
				included[d3mIndex] = filterIndex;
			} else {
				excluded[d3mIndex] = filterIndex;
			}
			d3mIndices[d3mIndex] = true;
		});
	});

	const includes = {
		type: ROW_FILTER,
		mode: INCLUDE_FILTER,
		d3mIndices: []
	};;
	const excludes = {
		type: ROW_FILTER,
		mode: EXCLUDE_FILTER,
		d3mIndices: []
	};

	_.keys(d3mIndices).forEach(d3mIndex => {
		const includedIndex = included[d3mIndex];
		const excludedIndex = excluded[d3mIndex];

		// NOTE: filters should be in the order they are created
		if (includedIndex >= 0 && excludedIndex >= 0) {
			// if excluded and then included, omit filter entirely
			return;
		}

		if (includedIndex >= 0) {
			includes.d3mIndices.push(d3mIndex);
			return;
		}

		if (excludedIndex >= 0) {
			excludes.d3mIndices.push(d3mIndex);
		}
	});

	if (includes.d3mIndices.length > 0) {
		remaining.push(includes);
	}

	if (excludes.d3mIndices.length > 0) {
		remaining.push(excludes);
	}

	return remaining;
}

function addFilter(filters: string, filter: Filter): string {
	const decoded = decodeFilters(filters);
	decoded.push(filter);
	return encodeFilters(dedupeRowFilters(decoded));
}

function removeFilter(filters: string, filter: Filter): string {
	// decode the provided filters
	const decoded = decodeFilters(filters);
	const index = _.findIndex(decoded, f => {
		return _.isEqual(f, filter);
	});
	if (index !== -1) {
		decoded.splice(index, 1);
	}
	// encode the filters back into a url string
	return encodeFilters(decoded);
}

export function addFilterToRoute(component: Vue, filter: Filter) {
	// retrieve the filters from the route
	const filters = routeGetters.getRouteFilters(component.$store);
	// merge the updated filters back into the route query params
	const updated = addFilter(filters, filter);
	const entry = overlayRouteEntry(routeGetters.getRoute(component.$store), {
		filters: updated
	});
	component.$router.push(entry);
}

export function removeFilterFromRoute(component: Vue, filter: Filter) {
	// retrieve the filters from the route
	const filters = routeGetters.getRouteFilters(component.$store);
	// merge the updated filters back into the route query params
	const updated = removeFilter(filters, filter);
	const entry = overlayRouteEntry(routeGetters.getRoute(component.$store), {
		filters: updated
	});
	component.$router.push(entry);
}

export function removeFiltersByName(component: Vue, name: string) {
	// retrieve the filters from the route
	const filters = routeGetters.getRouteFilters(component.$store);
	let decoded = decodeFilters(filters);
	decoded = decoded.filter(filter => {
		return (filter.name !== name);
	});
	const encoded = encodeFilters(decoded);
	const entry = overlayRouteEntry(routeGetters.getRoute(component.$store), {
		filters: encoded
	});
	component.$router.push(entry);
}
