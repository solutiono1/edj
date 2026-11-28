import { initiateDappRelay } from '@wizardconnect/core';
import { DappConnectionManager, loadSession } from '@wizardconnect/dapp';
import {
  hash160,
  encodeCashAddress,
  cashAddressToLockingBytecode,
  lockingBytecodeToCashAddress,
  encodeTransaction,
  decodeTransaction,
  binToHex,
  hexToBin,
  sha256,
  hash256,
  secp256k1,
  decodePrivateKeyWif,
  encodePrivateKeyWif,
  encodeDataPush,
  encodeSigningSerializationBch,
  SigningSerializationType,
  numberToBinUint32LE,
  valueSatoshisToBin,
  encodeTransactionOutput,
  encodeTransactionOutpoints,
  encodeTransactionInputSequenceNumbersForSigning,
  flattenBinArray
} from '@bitauth/libauth';
import QRCode from 'qrcode';

function pubkeyToCashAddress(pubkey, prefix = 'bitcoincash') {
  return encodeCashAddress({ prefix, type: 'p2pkh', payload: hash160(pubkey) }).address;
}

async function renderQR(uri) {
  return await QRCode.toDataURL(uri, { margin: 1, width: 320, errorCorrectionLevel: 'M' });
}

export {
  initiateDappRelay,
  DappConnectionManager,
  loadSession,
  pubkeyToCashAddress,
  renderQR,
  cashAddressToLockingBytecode,
  lockingBytecodeToCashAddress,
  encodeTransaction,
  decodeTransaction,
  binToHex,
  hexToBin,
  sha256,
  hash160,
  hash256,
  secp256k1,
  decodePrivateKeyWif,
  encodePrivateKeyWif,
  encodeDataPush,
  encodeSigningSerializationBch,
  SigningSerializationType,
  numberToBinUint32LE,
  valueSatoshisToBin,
  encodeTransactionOutput,
  encodeTransactionOutpoints,
  encodeTransactionInputSequenceNumbersForSigning,
  flattenBinArray
};
