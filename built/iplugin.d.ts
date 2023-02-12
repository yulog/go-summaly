/// <reference types="node" />
import * as URL from 'node:url';
import Summary from './summary.js';
export interface IPlugin {
    test: (url: URL.Url) => boolean;
    summarize: (url: URL.Url, lang?: string) => Promise<Summary>;
}
