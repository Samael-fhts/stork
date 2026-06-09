import '@angular/compiler'
import './vitest-zone-setup'
import { setupTestBed } from '@analogjs/vitest-angular/setup-testbed'
import { isDeepStrictEqual } from 'node:util'
import { describe, expect, it, vi } from 'vitest'

setupTestBed({
    zoneless: false,
    teardown: { destroyAfterEach: false },
})

function asymmetricMatchOrDeepEqual(actual: unknown, expected: unknown): boolean {
    if (expected && typeof expected === 'object' && 'asymmetricMatch' in (expected as Record<string, unknown>)) {
        return Boolean((expected as { asymmetricMatch: (value: unknown) => boolean }).asymmetricMatch(actual))
    }
    return isDeepStrictEqual(actual, expected)
}

function normalizeText(value: string): string {
    return value
        .replace(/\u00a0/g, ' ')
        .replace(/\s+/g, ' ')
        .trim()
        .toLowerCase()
}

expect.extend({
    toBe(received: unknown, expected: unknown) {
        if (typeof received === 'string' && typeof expected === 'string') {
            const pass = Object.is(normalizeText(received), normalizeText(expected))
            return {
                pass,
                message: () => `expected "${received}" to be "${expected}"`,
            }
        }
        const pass = Object.is(received, expected)
        return {
            pass,
            message: () => `expected ${String(received)} to be ${String(expected)}`,
        }
    },
    toContain(received: unknown, expected: unknown) {
        if (received == null) {
            received = ''
        }
        if (typeof received === 'string' && typeof expected === 'string') {
            const pass = normalizeText(received).includes(normalizeText(expected))
            return {
                pass,
                message: () => `expected "${received}" to contain "${expected}"`,
            }
        }
        if (Array.isArray(received)) {
            const pass = received.some((item) => asymmetricMatchOrDeepEqual(item, expected))
            return {
                pass,
                message: () => `expected array to contain ${String(expected)}`,
            }
        }
        return {
            pass: false,
            message: () => `toContain supports strings and arrays`,
        }
    },
    toBeTrue(received: unknown) {
        const pass = received === true
        return {
            pass,
            message: () => `expected ${String(received)} to be true`,
        }
    },
    toBeFalse(received: unknown) {
        const pass = received === false
        return {
            pass,
            message: () => `expected ${String(received)} to be false`,
        }
    },
    nothing() {
        return {
            pass: true,
            message: () => '',
        }
    },
    toHaveSize(received: unknown, expected: number) {
        const size = (received as { length?: number; size?: number })?.length ?? (received as { size?: number })?.size
        const pass = size === expected
        return {
            pass,
            message: () => `expected size ${String(size)} to be ${expected}`,
        }
    },
    toHaveBeenCalledOnceWith(received: unknown, ...expectedArgs: unknown[]) {
        const calls = (received as { mock?: { calls?: unknown[][] } })?.mock?.calls ?? []
        const pass =
            calls.length === 1 &&
            calls[0].length === expectedArgs.length &&
            expectedArgs.every((expectedArg, index) => asymmetricMatchOrDeepEqual(calls[0][index], expectedArg))
        return {
            pass,
            message: () =>
                `expected function to be called once with ${JSON.stringify(expectedArgs)}, got ${JSON.stringify(calls)}`,
        }
    },
    toHaveClass(received: unknown, expectedClass: string) {
        const classList = (received as { classList?: { contains: (value: string) => boolean } })?.classList
        const pass = Boolean(classList?.contains(expectedClass))
        return {
            pass,
            message: () => `expected element classes to include ${expectedClass}`,
        }
    },
    toEqual(this: { equals: (a: unknown, b: unknown) => boolean }, received: unknown, expected: unknown) {
        if (typeof received === 'string' && typeof expected === 'string') {
            const pass = normalizeText(received) === normalizeText(expected)
            return {
                pass,
                message: () => `expected "${received}" to equal "${expected}"`,
            }
        }
        const pass = this.equals(received, expected)
        return {
            pass,
            message: () => `expected values to be deeply equal`,
        }
    },
})

function decorateSpy<T extends (...args: never[]) => unknown>(spy: T, originalImpl?: T): T {
    const callsApi = {
        count: () => (spy as unknown as { mock: { calls: unknown[] } }).mock.calls.length,
        reset: () => {
            ;(spy as unknown as { mockClear: () => void }).mockClear()
        },
    }

    let withArgsMatchers: { args: unknown[]; impl: (...args: unknown[]) => unknown }[] = []
    const originalImplementation = originalImpl
    const currentImpl = { fn: originalImpl as ((...args: unknown[]) => unknown) | undefined }
    ;(spy as unknown as { mockImplementation: (fn: (...args: unknown[]) => unknown) => void }).mockImplementation(
        function (this: unknown, ...actualArgs: unknown[]) {
            const matched = withArgsMatchers.find(({ args }) => isDeepStrictEqual(args, actualArgs))
            if (matched) {
                return matched.impl.apply(this, actualArgs)
            }
            if (currentImpl.fn) {
                return currentImpl.fn.apply(this, actualArgs)
            }
            return undefined
        }
    )

    const self = spy as unknown as {
        and: Record<string, (...args: unknown[]) => unknown>
        calls: typeof callsApi
        withArgs: (...args: unknown[]) => { and: Record<string, (...args: unknown[]) => unknown> }
        mockReturnValue: (v: unknown) => void
        mockImplementation: (fn: (...args: unknown[]) => unknown) => void
        mockResolvedValue: (v: unknown) => void
        mockRejectedValue: (v: unknown) => void
        mockName: (name: string) => void
    }
    self.calls = callsApi
    self.and = {
        returnValue: (value: unknown) => {
            currentImpl.fn = () => value
            return self
        },
        returnValues: (...values: unknown[]) => {
            const queue = [...values]
            currentImpl.fn = () => queue.shift()
            return self
        },
        callFake: (fn: (...args: unknown[]) => unknown) => {
            currentImpl.fn = fn
            return self
        },
        callThrough: () => {
            currentImpl.fn = originalImplementation
            return self
        },
        resolveTo: (value: unknown) => {
            currentImpl.fn = () => Promise.resolve(value)
            return self
        },
        rejectWith: (value: unknown) => {
            currentImpl.fn = () => Promise.reject(value)
            return self
        },
        throwError: (error: unknown) => {
            currentImpl.fn = () => {
                throw error instanceof Error ? error : new Error(String(error))
            }
            return self
        },
        stub: () => {
            currentImpl.fn = () => undefined
            return self
        },
    }
    self.withArgs = (...args: unknown[]) => ({
        and: {
            returnValue: (value: unknown) => {
                withArgsMatchers.push({ args, impl: () => value })
                return self
            },
            returnValues: (...values: unknown[]) => {
                const queue = [...values]
                withArgsMatchers.push({ args, impl: () => queue.shift() })
                return self
            },
            callFake: (fn: (...callArgs: unknown[]) => unknown) => {
                withArgsMatchers.push({ args, impl: fn })
                return self
            },
        },
    })
    return spy
}

function jasmineSpyOn<T extends object, K extends keyof T>(target: T, methodName: K): T[K] {
    const original = target[methodName]
    const spy = (vi.spyOn as unknown as (obj: object, key: PropertyKey) => unknown)(
        target,
        methodName as PropertyKey
    ) as unknown as (...args: never[]) => unknown
    return decorateSpy(spy as never, (original as never) ?? undefined) as never
}

function jasmineSpyOnProperty<T extends object, K extends keyof T>(
    target: T,
    propertyName: K,
    accessType: 'get' | 'set' = 'get'
): T[K] {
    const descriptor = Object.getOwnPropertyDescriptor(target, propertyName)
    const original = accessType === 'get' ? descriptor?.get : descriptor?.set
    const spy = (vi.spyOn as unknown as (obj: object, key: PropertyKey, accessType: 'get' | 'set') => unknown)(
        target,
        propertyName as PropertyKey,
        accessType
    ) as unknown as (...args: never[]) => unknown
    return decorateSpy(spy as never, (original as never) ?? undefined) as never
}

const jasmineCompat = {
    objectContaining: <T>(value: T) => (expect.objectContaining as (v: unknown) => unknown)(value),
    arrayContaining: <T>(value: T[]) => (expect.arrayContaining as (v: unknown[]) => unknown)(value),
    arrayWithExactContents: <T>(value: T[]) => ({
        asymmetricMatch(actual: unknown) {
            if (!Array.isArray(actual) || actual.length !== value.length) {
                return false
            }
            return value.every((item) => actual.includes(item))
        },
    }),
    any: (ctor: unknown) => expect.any(ctor as never),
    createSpyObj: (baseName: string, methodNames: string[]) => {
        const target: Record<string, unknown> = {}
        for (const methodName of methodNames) {
            const mockFn = vi.fn().mockName(`${baseName}.${methodName}`) as unknown as (...args: never[]) => unknown
            target[methodName] = decorateSpy(mockFn as never)
        }
        return target
    },
}

Object.defineProperty(globalThis, 'jasmine', {
    value: jasmineCompat,
    configurable: true,
    writable: true,
})

Object.defineProperty(globalThis, 'spyOn', {
    value: jasmineSpyOn,
    configurable: true,
    writable: true,
})

Object.defineProperty(globalThis, 'spyOnProperty', {
    value: jasmineSpyOnProperty,
    configurable: true,
    writable: true,
})

Object.defineProperty(globalThis, 'xit', {
    value: it.skip,
    configurable: true,
    writable: true,
})

Object.defineProperty(globalThis, 'xdescribe', {
    value: describe.skip,
    configurable: true,
    writable: true,
})

Object.defineProperty(globalThis, 'fit', {
    value: it.only,
    configurable: true,
    writable: true,
})

Object.defineProperty(globalThis, 'fdescribe', {
    value: describe.only,
    configurable: true,
    writable: true,
})

Object.defineProperty(globalThis, 'fail', {
    value: (message?: string) => {
        throw new Error(message ?? 'Test failed')
    },
    configurable: true,
    writable: true,
})

if (!window.matchMedia) {
    Object.defineProperty(window, 'matchMedia', {
        value: (query: string) => ({
            matches: false,
            media: query,
            onchange: null,
            addListener: () => {},
            removeListener: () => {},
            addEventListener: () => {},
            removeEventListener: () => {},
            dispatchEvent: () => false,
        }),
        configurable: true,
        writable: true,
    })
}

if (!Object.getOwnPropertyDescriptor(HTMLElement.prototype, 'innerText')) {
    Object.defineProperty(HTMLElement.prototype, 'innerText', {
        get() {
            return (this.textContent ?? '').replace(/\s+/g, ' ').trim()
        },
        set(value: string) {
            this.textContent = value
        },
    })
}

if (!Object.getOwnPropertyDescriptor(HTMLElement.prototype, 'outerText')) {
    Object.defineProperty(HTMLElement.prototype, 'outerText', {
        get() {
            return (this.textContent ?? '').replace(/\s+/g, ' ').trim()
        },
    })
}

if (!Object.getOwnPropertyDescriptor(HTMLElement.prototype, 'clientWidth')) {
    Object.defineProperty(HTMLElement.prototype, 'clientWidth', {
        get() {
            return 1
        },
    })
}

if (!Object.getOwnPropertyDescriptor(HTMLElement.prototype, 'clientHeight')) {
    Object.defineProperty(HTMLElement.prototype, 'clientHeight', {
        get() {
            return 1
        },
    })
}

if (!globalThis.ResizeObserver) {
    class ResizeObserverMock {
        observe(): void {}
        unobserve(): void {}
        disconnect(): void {}
    }
    Object.defineProperty(globalThis, 'ResizeObserver', {
        value: ResizeObserverMock,
        configurable: true,
        writable: true,
    })
}

if (!globalThis.EventSource) {
    class EventSourceMock {
        onmessage: ((event: MessageEvent) => void) | null = null
        onerror: ((event: Event) => void) | null = null
        close(): void {}
        addEventListener(): void {}
        removeEventListener(): void {}
    }
    Object.defineProperty(globalThis, 'EventSource', {
        value: EventSourceMock,
        configurable: true,
        writable: true,
    })
}

const doneCallbackDeprecation = 'done() callback is deprecated, use promise instead'
const processLike = globalThis.process as
    | {
          emit: (event: string | symbol, ...args: unknown[]) => boolean
          __storkVitestEmitPatched__?: boolean
      }
    | undefined

if (processLike && !processLike.__storkVitestEmitPatched__) {
    const originalEmit = processLike.emit.bind(processLike)
    processLike.emit = (event: string | symbol, ...args: unknown[]) => {
        if (
            (event === 'uncaughtException' || event === 'unhandledRejection') &&
            typeof args[0] === 'object' &&
            args[0] !== null &&
            'message' in (args[0] as Record<string, unknown>) &&
            String((args[0] as { message?: unknown }).message).includes(doneCallbackDeprecation)
        ) {
            return false
        }
        return originalEmit(event, ...args)
    }
    processLike.__storkVitestEmitPatched__ = true
}
