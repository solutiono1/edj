import { randomBytes } from 'node:crypto'
import { WIF, TEST_NETWORK } from '@scure/btc-signer'
import { SignatureTemplate } from 'cashscript'
import { hex } from '@scure/base'
import { hash160 } from '@cashscript/utils'
import { Address, TEST_NETWORK as CASHADDR_TEST } from 'cashaddr'

const priv = randomBytes(32)
const pub = new SignatureTemplate(priv).getPublicKey()
const pkh = hex.encode(hash160(pub))
const address = Address(CASHADDR_TEST).encode({ format: 'cashaddr', type: 'pkh', hash: hex.decode(pkh) })
const wif = WIF(TEST_NETWORK).encode(priv)

console.log('CHIPNET_WIF=' + wif)
console.log('ADDRESS=' + address)
console.log('PKH=' + pkh)
console.log('')
console.log('Fund the address from https://tbch.googol.cash (choose chipnet), then use these values.')
console.log('PAYER_PKH is the same PKH. PAYEE_PKH can be a different wallet you control, or the same PKH for a test.')
console.log('REFUND_AFTER should be the current chipnet block height plus a few blocks.')
