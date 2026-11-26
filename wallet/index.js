import { initiateDappRelay } from '@wizardconnect/core';
import { DappConnectionManager, loadSession } from '@wizardconnect/dapp';
import {
  hash160,
  encodeCashAddress,
  cashAddressToLockingBytecode,
  lockingBytecodeToCashAddress,
  encodeTransaction,
  binToHex,
  hexToBin,
  sha256
} from '@bitauth/libauth';
import QRCode from 'qrcode';

function pubkeyToCashAddress(pubkey, prefix = 'bitcoincash') {
  return encodeCashAddress(prefix, 'p2pkh', hash160(pubkey));
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
  binToHex,
  hexToBin,
  sha256
};
