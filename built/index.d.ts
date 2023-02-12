/**
 * summaly
 * https://github.com/syuilo/summaly
 */
import Summary from './summary.js';
import type { IPlugin as _IPlugin } from './iplugin.js';
export declare type IPlugin = _IPlugin;
import * as Got from 'got';
import type { FastifyInstance } from 'fastify';
declare type Options = {
    /**
     * Accept-Language for the request
     */
    lang?: string | null;
    /**
     * Whether follow redirects
     */
    followRedirects?: boolean;
    /**
     * Custom Plugins
     */
    plugins?: IPlugin[];
    /**
     * Custom HTTP agent
     */
    agent?: Got.Agents;
};
declare type Result = Summary & {
    /**
     * The actual url of that web page
     */
    url: string;
};
/**
 * Summarize an web page
 */
export declare const summaly: (url: string, options?: Options | undefined) => Promise<Result>;
export default function (fastify: FastifyInstance, options: Options, done: (err?: Error) => void): void;
export {};
