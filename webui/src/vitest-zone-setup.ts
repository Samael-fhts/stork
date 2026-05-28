import 'zone.js'
import 'zone.js/plugins/sync-test'
import 'zone.js/plugins/proxy'
import 'zone.js/testing'
import '@analogjs/vitest-angular/setup-snapshots'

const ZoneCtor = (globalThis as Record<string, unknown>)['Zone'] as {
    current: { fork: (spec: unknown) => { run: (fn: (...args: unknown[]) => unknown, thisArg: unknown, args?: unknown[]) => unknown } }
    SyncTestZoneSpec: new (name: string) => unknown
    ProxyZoneSpec: new () => unknown
}

if (!ZoneCtor) throw new Error('Missing: Zone (zone.js)')
if ((globalThis as Record<string, unknown>)['__vitest_zone_patch__'] === true) {
    throw new Error("'vitest' has already been patched with 'Zone'.")
}
(globalThis as Record<string, unknown>)['__vitest_zone_patch__'] = true

const syncZone = ZoneCtor.current.fork(new ZoneCtor.SyncTestZoneSpec('vitest.describe'))
const testProxyZone = ZoneCtor.current.fork(new ZoneCtor.ProxyZoneSpec())

function wrapDescribeInZone(describeBody?: (...args: unknown[]) => unknown) {
    if (!describeBody) return describeBody
    return function (...args: unknown[]) {
        return syncZone.run(describeBody, null, args)
    }
}

function wrapTestInZone(testBody?: (...args: unknown[]) => unknown) {
    if (!testBody) return testBody
    // Always expose arity 0 to Vitest to avoid done()-style callback mode.
    return function (...args: unknown[]) {
        return testProxyZone.run(testBody, null, args)
    }
}

function patchDescribe(name: 'describe') {
    const env = globalThis as Record<string, unknown>
    const original = env[name] as (...args: unknown[]) => unknown
    env[name] = function (...args: unknown[]) {
        args[1] = wrapDescribeInZone(args[1] as (...args: unknown[]) => unknown)
        return original.apply(this, args)
    }
    const patched = env[name] as Record<string, (...args: unknown[]) => unknown>
    const origAny = original as unknown as Record<string, (...args: unknown[]) => unknown>
    patched.each = (...eachArgs: unknown[]) => (...args: unknown[]) => {
        args[1] = wrapDescribeInZone(args[1] as (...args: unknown[]) => unknown)
        return origAny.each.apply(original, eachArgs).apply(original, args)
    }
    patched.only = (...eachArgs: unknown[]) => (...args: unknown[]) => {
        args[1] = wrapDescribeInZone(args[1] as (...args: unknown[]) => unknown)
        return origAny.only.apply(original, eachArgs).apply(original, args)
    }
    patched.skip = (...eachArgs: unknown[]) => (...args: unknown[]) => {
        args[1] = wrapDescribeInZone(args[1] as (...args: unknown[]) => unknown)
        return origAny.skip.apply(original, eachArgs).apply(original, args)
    }
}

function patchTestLike(name: 'test' | 'it' | 'beforeEach' | 'afterEach' | 'beforeAll' | 'afterAll') {
    const env = globalThis as Record<string, unknown>
    const original = env[name] as (...args: unknown[]) => unknown
    env[name] = function (...args: unknown[]) {
        const idx = name.startsWith('before') || name.startsWith('after') ? 0 : 1
        args[idx] = wrapTestInZone(args[idx] as (...args: unknown[]) => unknown)
        return original.apply(this, args)
    }
    const patched = env[name] as Record<string, (...args: unknown[]) => unknown>
    const origAny = original as unknown as Record<string, (...args: unknown[]) => unknown>
    if (origAny.each) {
        patched.each = (...eachArgs: unknown[]) => (...args: unknown[]) => {
            args[1] = wrapTestInZone(args[1] as (...args: unknown[]) => unknown)
            return origAny.each.apply(original, eachArgs).apply(original, args)
        }
    }
    if (origAny.only) {
        patched.only = (...eachArgs: unknown[]) => (...args: unknown[]) => {
            args[1] = wrapTestInZone(args[1] as (...args: unknown[]) => unknown)
            return origAny.only.apply(original, eachArgs).apply(original, args)
        }
    }
    if (origAny.skip) {
        patched.skip = (...eachArgs: unknown[]) => (...args: unknown[]) => {
            args[1] = wrapTestInZone(args[1] as (...args: unknown[]) => unknown)
            return origAny.skip.apply(original, eachArgs).apply(original, args)
        }
    }
}

patchDescribe('describe')
patchTestLike('test')
patchTestLike('it')
patchTestLike('beforeEach')
patchTestLike('afterEach')
patchTestLike('beforeAll')
patchTestLike('afterAll')
