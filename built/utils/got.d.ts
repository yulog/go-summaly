import * as Got from 'got';
import * as cheerio from 'cheerio';
export declare let agent: Got.Agents;
export declare function setAgent(_agent: Got.Agents): void;
export declare type GotOptions = {
    url: string;
    method: 'GET' | 'POST' | 'HEAD';
    body?: string;
    headers: Record<string, string | undefined>;
    typeFilter?: RegExp;
};
export declare function scpaping(url: string, opts?: {
    lang?: string;
}): Promise<{
    body: string;
    $: cheerio.CheerioAPI;
    response: Got.Response<string>;
}>;
export declare function get(url: string): Promise<string>;
export declare function head(url: string): Promise<Got.Response<string>>;
