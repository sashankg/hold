import { createLibp2p } from "libp2p";
import { webSockets } from '@libp2p/websockets'
import { noise } from '@chainsafe/libp2p-noise'
import { RELAY_V2_HOP_CODEC, circuitRelayTransport } from '@libp2p/circuit-relay-v2'
import { FetchComponents, http } from "@libp2p/http-fetch";
import { peerIdFromString } from '@libp2p/peer-id'
import { multiaddr } from "@multiformats/multiaddr";
import { yamux } from '@chainsafe/libp2p-yamux'
import { pipe } from 'it-pipe'
import * as websocketFilters from '@libp2p/websockets/filters'


localStorage.setItem('debug', '*libp2p:*')

const p2pURIPrefix = "p2p:"

const node = await createLibp2p({
  streamMuxers: [
    // mplex(),
    yamux(),
  ],
  transports: [
    webSockets({
      filter: websocketFilters.all,
    }),
    circuitRelayTransport({
      discoverRelays: 0,
    })
  ],
  connectionEncryption: [noise()],
  connectionGater: {
    denyDialMultiaddr: () => false,
  },
  services: {
    httpFetch: (components: FetchComponents) => {
      const httpService = http()(components)
      return (req: Request, init?: RequestInit) => {
        if (req.url.startsWith(p2pURIPrefix)) {
          req.arrayBuffer
        }
      }
    }
  },
  logger: {
    forComponent: (component) => {
      const logger = (formatter: any, ...args: any[]) => console.log(component, formatter, ...args)
      logger.error = console.error
      logger.trace = () => { }
      logger.enabled = true
      return logger
    }
  },
  peerRouters: [
    (components) => {
      return {
        findPeer: async (id, options) => {
          return {
            id,
            multiaddrs: [
              multiaddr("/ip4/127.0.0.1/tcp/4002/ws/p2p/QmNpBvAKWrjigDHP4Mn3LpqCmin5F2K9TiVFoFGTC6ayV3/p2p-circuit")
            ]
          }
        },
        getClosestPeers: async function*() { }
      }
    }
  ],
  connectionManager: {
    minConnections: 0,
  }
})

const relay = await node.peerStore.save(peerIdFromString("QmNpBvAKWrjigDHP4Mn3LpqCmin5F2K9TiVFoFGTC6ayV3"), {
  protocols: [RELAY_V2_HOP_CODEC],
  addresses: [
    {
      multiaddr: multiaddr("/ip4/127.0.0.1/tcp/4002/ws/"),
      isCertified: true
    }
  ]
})

console.log(relay)

// const conn = await node.dial(relay.id)

// const stream = await conn.newStream("/ipfs/id/1.0.0")

// pipe(stream.source, async (x) => {
//   let next: IteratorResult<any> = { done: false, value: null }
//   while (!next.done) {
//     next = await x.next()
//     if (next.value) {
//       console.log(next.value)
//     }
//   }
// })

// stream.close()
//

class TestHeaders implements Headers {
  map: Map<string, string>;
  constructor(init?: Record<string, string>) {
    this.map = new Map()
    if (init) {
      for (const key in init) {
        this.map.set(key, init[key])
      }
    }
  }
  append(name: string, value: string): void {
    this.map.set(name, value)
  }
  delete(name: string): void {
    this.map.delete(name)
  }
  get(name: string): string | null {
    return this.map.get(name) ?? null
  }
  getSetCookie(): string[] {
    return []
  }
  has(name: string): boolean {
    return this.map.has(name)
  }
  set(name: string, value: string): void {
    this.map.set(name, value)
  }
  forEach(callbackfn: (value: string, key: string, parent: TestHeaders) => void, thisArg?: any): void {
    return this.map.forEach((value, key) => {
      return callbackfn(value, key, this)
    }, thisArg)
  }
  entries(): IterableIterator<[string, string]> {
    return this.map.entries()
  }
  keys(): IterableIterator<string> {
    return this.map.keys()
  }
  values(): IterableIterator<string> {
    return this.map.values()
  }
  [Symbol.iterator](): IterableIterator<[string, string]> {
    return this.map[Symbol.iterator]()
  }
}
class TestRequest implements Request {
  _request: Request;
  _headers: TestHeaders
  constructor(url: string, init?: Omit<RequestInit, 'headers'> & { headers: Record<string, string> }) {
    this._request = new Request(url, init)
    this._headers = new TestHeaders(init?.headers)
  }
  get cache(): RequestCache {
    return this._request.cache
  };
  get credentials(): RequestCredentials {
    return this._request.credentials
  };
  get destination(): RequestDestination {
    return this._request.destination
  };
  get headers(): Headers {
    return this._headers
  };
  get integrity(): string {
    return this._request.integrity
  };
  get keepalive(): boolean {
    return this._request.keepalive
  };
  get method(): string {
    return this._request.method
  };
  get mode(): RequestMode {
    return this._request.mode
  };
  get redirect(): RequestRedirect {
    return this._request.redirect
  };
  get referrer(): string {
    return this._request.referrer
  };
  get referrerPolicy(): ReferrerPolicy {
    return this._request.referrerPolicy
  };
  get signal(): AbortSignal {
    return this._request.signal
  };
  get url(): string {
    return this._request.url
  };
  clone(): Request {
    return this._request.clone()
  }
  get body(): ReadableStream<Uint8Array> | null {
    return this._request.body
  }
  get bodyUsed(): boolean {
    return this._request.bodyUsed
  };
  arrayBuffer(): Promise<ArrayBuffer> {
    return this._request.arrayBuffer()
  }
  blob(): Promise<Blob> {
    return this._request.blob()
  }
  formData(): Promise<FormData> {
    return this._request.formData()
  }
  json(): Promise<any> {
    return this._request.json()
  }
  text(): Promise<string> {
    return this._request.text()
  }
}

const resp = await node.services.http.fetch(
  new TestRequest("multiaddr:/p2p/QmShjXauvefhMYJ7tfam4adsMTy1DaxJBTKQEVh6Bs3m5i/http-path/graph", {
    headers: {
      "Host": "somthing:test",
      "Content-Type": "application/json"
    },
  })
)


console.log(await resp.text())

await node.dial(peerIdFromString("QmShjXauvefhMYJ7tfam4adsMTy1DaxJBTKQEVh6Bs3m5i"))

console.log(node)
