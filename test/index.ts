/**
 * Tests!
 */

'use strict';

/* dependencies below */

import * as fs from 'fs';
import fastify from 'fastify';
import { summaly } from '../src/index.js';
import { dirname } from 'node:path';
import { fileURLToPath } from 'node:url';
import {expect, jest, test, describe, beforeEach, afterEach} from '@jest/globals';

const _filename = fileURLToPath(import.meta.url);
const _dirname = dirname(_filename);

/* settings below */

Error.stackTraceLimit = Infinity;

// During the test the env variable is set to test
process.env.NODE_ENV = 'test';
process.env.SUMMALY_ALLOW_PRIVATE_IP = 'true';

const port = 3060;
const host = `http://localhost:${port}`;

// Display detail of unhandled promise rejection
process.on('unhandledRejection', console.dir);

let app: ReturnType<typeof fastify>;

afterEach(() => {
	if (app) return app.close();
});

/* tests below */

test('faviconがHTML上で指定されていないが、ルートに存在する場合、正しく設定される', async () => {
	app = fastify({
		logger: true,
	});
	app.get('/', (request, reply) => {
		return reply.send(fs.createReadStream(_dirname + '/htmls/no-favicon.html'));
	});
	app.get('/favicon.ico', (_, reply) => reply.status(200));
	await app.listen({ port });

	const summary = await summaly(host);
	expect(summary.icon).toBe(`${host}/favicon.ico`);
});

test('faviconがHTML上で指定されていなくて、ルートにも存在しなかった場合 null になる', async () => {
	app = fastify();
	app.get('/', (request, reply) => {
		return reply.send(fs.createReadStream(_dirname + '/htmls/no-favicon.html'));
	});
	await app.listen({ port });

	const summary = await summaly(host);
	expect(summary.icon).toBe(null);
});

test('titleがcleanupされる', async () => {
	app = fastify();
	app.get('/', (request, reply) => {
		return reply.send(fs.createReadStream(_dirname + '/htmls/ditry-title.html'));
	});
	await app.listen({ port });

	const summary = await summaly(host);
	expect(summary.title).toBe('Strawberry Pasta');
});

describe('Private IP blocking', () => {
	beforeEach(() => {
		process.env.SUMMALY_ALLOW_PRIVATE_IP = 'false';
	});

	test('private ipなサーバーの情報を取得できない', async () => {
		app = fastify();
		app.get('/', (request, reply) => {
			return reply.send(fs.createReadStream(_dirname + '/htmls/og-title.html'));
		});
		await app.listen({ port });


		expect(() => summaly(host)).toThrow();
	});

	afterEach(() => {
		process.env.SUMMALY_ALLOW_PRIVATE_IP = 'true';
	});
});

describe('OGP', () => {
	test('title', async () => {
		app = fastify();
		app.get('/', (request, reply) => {
			return reply.send(fs.createReadStream(_dirname + '/htmls/og-title.html'));
		});
		await app.listen({ port });

		const summary = await summaly(host);
		expect(summary.title).toBe('Strawberry Pasta');
	});

	test('description', async () => {
		app = fastify();
		app.get('/', (request, reply) => {
			return reply.send(fs.createReadStream(_dirname + '/htmls/og-description.html'));
		});
		await app.listen({ port });

		const summary = await summaly(host);
		expect(summary.description).toBe('Strawberry Pasta');
	});

	test('site_name', async () => {
		app = fastify();
		app.get('/', (request, reply) => {
			return reply.send(fs.createReadStream(_dirname + '/htmls/og-site_name.html'));
		});
		await app.listen({ port });

		const summary = await summaly(host);
		expect(summary.sitename).toBe('Strawberry Pasta');
	});

	test('thumbnail', async () => {
		app = fastify();
		app.get('/', (request, reply) => {
			return reply.send(fs.createReadStream(_dirname + '/htmls/og-image.html'));
		});
		await app.listen({ port });

		const summary = await summaly(host);
		expect(summary.thumbnail).toBe('https://himasaku.net/himasaku.png');
	});
});

describe('TwitterCard', () => {
	test('title', async () => {
		app = fastify();
		app.get('/', (request, reply) => {
			return reply.send(fs.createReadStream(_dirname + '/htmls/twitter-title.html'));
		});
		await app.listen({ port });

		const summary = await summaly(host);
		expect(summary.title).toBe('Strawberry Pasta');
	});

	test('description', async () => {
		app = fastify();
		app.get('/', (request, reply) => {
			return reply.send(fs.createReadStream(_dirname + '/htmls/twitter-description.html'));
		});
		await app.listen({ port });

		const summary = await summaly(host);
		expect(summary.description).toBe('Strawberry Pasta');
	});

	test('thumbnail', async () => {
		app = fastify();
		app.get('/', (request, reply) => {
			return reply.send(fs.createReadStream(_dirname + '/htmls/twitter-image.html'));
		});
		await app.listen({ port });

		const summary = await summaly(host);
		expect(summary.thumbnail).toBe('https://himasaku.net/himasaku.png');
	});

	test('Player detection - PeerTube:video => video', async () => {
		app = fastify();
		app.get('/', (request, reply) => {
			return reply.send(fs.createReadStream(_dirname + '/htmls/player-peertube-video.html'));
		});
		await app.listen({ port });

		const summary = await summaly(host);
		expect(summary.player.url).toBe('https://example.com/embedurl');
	});

	test('Player detection - Pleroma:video => video', async () => {
		app = fastify();
		app.get('/', (request, reply) => {
			return reply.send(fs.createReadStream(_dirname + '/htmls/player-pleroma-video.html'));
		});
		await app.listen({ port });

		const summary = await summaly(host);
		expect(summary.player.url).toBe('https://example.com/embedurl');
	});

	test('Player detection - Pleroma:image => image', async () => {
		app = fastify();
		app.get('/', (request, reply) => {
			return reply.send(fs.createReadStream(_dirname + '/htmls/player-pleroma-image.html'));
		});
		await app.listen({ port });

		const summary = await summaly(host);
		expect(summary.thumbnail).toBe('https://example.com/imageurl');
	});
});
