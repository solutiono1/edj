import { readFileSync } from 'node:fs'
import { Contract, ElectrumNetworkProvider, Network, SignatureTemplate, TransactionBuilder } from 'cashscript'
import { Address, TEST_NETWORK } from 'cashaddr'
import { hex } from '@scure/base'
import { hash160 } from '@cashscript/utils'

const artifact = JSON.parse(readFileSync(new URL('./PrivatePayroll.json', import.meta.url), 'utf8'))

const wif = process.env.CHIPNET_WIF
if (!wif) {
  console.error('Set CHIPNET_WIF to a funded chipnet WIF before deploying')
  process.exit(1)
}

const provider = new ElectrumNetworkProvider(Network.CHIPNET)
const amount = BigInt(process.env.CONTRACT_SATS || '25000')
let refundAfter = BigInt(process.env.REFUND_AFTER || '0')
const payeePkh = process.env.PAYEE_PKH
const payerPkh = process.env.PAYER_PKH

if (!payeePkh || !payerPkh) {
  console.error('Set PAYEE_PKH and PAYER_PKH as 20-byte hex strings')
  process.exit(1)
}
if (refundAfter === 0n) {
  refundAfter = BigInt((await provider.getBlockHeight()) + 10)
}

const template = new SignatureTemplate(wif)
const ownerPkh = hex.encode(hash160(template.getPublicKey()))
const ownerAddress = Address(TEST_NETWORK).encode({
  format: 'cashaddr',
  type: 'pkh',
  hash: hex.decode(ownerPkh)
})

const contract = new Contract(artifact, [payeePkh, payerPkh, amount, refundAfter], { provider })
const utxos = await provider.getUtxos(ownerAddress)
const usable = utxos
  .filter((utxo) => utxo.satoshis > amount + 1000n)
  .sort((a, b) => Number(a.satoshis - b.satoshis))

if (!usable.length) {
  console.error(`No funded chipnet UTXO at ${ownerAddress} above ${amount + 1000n} sats`)
  process.exit(1)
}

const builder = new TransactionBuilder({ provider })
builder.addInput(usable[0], template.unlockP2PKH())
builder.addOutput({ to: contract.address, amount })

if (process.env.DRY_RUN === '1') {
  console.log('dry run')
  console.log('owner', ownerAddress)
  console.log('contract', contract.address)
  console.log('input', usable[0].txid, usable[0].vout, usable[0].satoshis)
  console.log('raw', builder.build())
} else {
  const result = await builder.send()
  console.log('txid', result.txid)
  console.log('contract', contract.address)
  console.log('owner', ownerAddress)
}
