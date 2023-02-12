/// <reference types="node" />
import * as URL from 'node:url';
import summary from '../summary.js';
export declare function test(url: URL.Url): boolean;
export declare function summarize(url: URL.Url): Promise<summary>;
