import React from 'react'
import {
  getContractCreationTx,
  getERC20Balance,
  getERC20Holders,
  getERC721Holders,
  getERC721Tokens,
  getERC721TokensForAccount,
  getEvents,
  getHolderForERC721Token,
  getInternalToTxs,
  getReportData,
  getToTxs
} from './client/fetcher'
import { TransactionHeader, TransactionRowItem, TransactionTable } from './components/table/Transactions'
import { EventHeader, EventRowItem, EventTable } from './components/table/Events'
import { ReportHeader, ReportRowItem, ReportTable } from './components/table/Report'
import {
  BalanceHeader,
  BalanceRowItem, TokenBalanceTable,
  TokenHeader,
  TokenHolderHeader,
  TokenHolderRowItem, TokenHolderTable,
  TokenRowItem, TokenTable
} from './components/table/Tokens'

export const HomePageId = 'Home'
export const ContractPageId = 'Contract'
export const ReportPageId = 'Report'

export const ContractCreationTx = {
  label: 'Contract Creation Tx',
  value: 'ContractCreationTx',
  fields: {},
  View: TransactionTable,
  getItems: (rpcEndpoint, params, options) => getContractCreationTx(rpcEndpoint, params.address),
}

export const ToTxs = {
  label: 'Transactions To Contract',
  value: 'ToTxs',
  fields: {},
  View: TransactionTable,
  getItems: (rpcEndpoint, params, options) => getToTxs(rpcEndpoint, params.address, options),
}

export const InternalToTxs = {
  label: 'Internal Transactions to Contract',
  value: 'InternalToTxs',
  fields: {},
  View: TransactionTable,
  getItems: (rpcEndpoint, params, options) => getInternalToTxs(rpcEndpoint, params.address, options),
}

export const Events = {
  label: 'Contract Events',
  value: 'Events',
  fields: {},
  View: EventTable,
  getItems: (rpcEndpoint, params, options) => getEvents(rpcEndpoint, params.address, options),
}

export const GenerateReport = {
  label: 'Full Report',
  value: 'GenerateReport',
  fields: {
    startBlock: 'required',
    endBlock: 'required',
  },
  View: ReportTable,
  getItems: (rpcEndpoint, params, options) => getReportData(rpcEndpoint, params.address, params.startBlockNumber, params.endBlockNumber, options),
}

export const ERC20TokenHolders = {
  label: 'ERC20 Token Holders',
  value: 'ERC20TokenHolders',
  fields: {
    block: 'required',
  },
  View: TokenHolderTable,
  getItems: (rpcEndpoint, params, options) => getERC20Holders(rpcEndpoint, params.address, params.atBlock, options),
}

export const ERC20TokenBalance = {
  label: 'ERC20 Token Balance',
  value: 'ERC20TokenBalance',
  fields: {
    account: 'required',
    startBlock: 'required',
    endBlock: 'required',
  },
  View: TokenBalanceTable,
  getItems: (rpcEndpoint, params, options) => getERC20Balance(rpcEndpoint, params.address, params.account, params.startBlockNumber, params.endBlockNumber, options),
}

export const ERC721Holders = {
  label: 'ERC721 Token Holders',
  value: 'ERC721Holders',
  fields: {
    block: 'required',
  },
  View: TokenHolderTable,
  getItems: (rpcEndpoint, params, options) => getERC721Holders(rpcEndpoint, params.address, params.atBlock, options),
}

export const ERC721Tokens = {
  label: 'ERC721 Tokens',
  value: 'ERC721Tokens',
  fields: {
    block: 'required',
  },
  View: TokenTable,
  getItems: (rpcEndpoint, params, options) => getERC721Tokens(rpcEndpoint, params.address, params.atBlock, options),
}

export const ERC721TokensForAccount = {
  label: 'ERC721 Tokens for Account',
  value: 'ERC721TokensForAccount',
  fields: {
    account: 'required',
    block: 'required',
  },
  View: TokenTable,
  getItems: (rpcEndpoint, params, options) => getERC721TokensForAccount(rpcEndpoint, params.address, params.account, params.atBlock, options),
}

export const ERC721HolderForToken = {
  label: 'Holder for ERC721',
  value: 'ERC721HolderForToken',
  fields: {
    tokenId: 'required',
    block: 'required',
  },
  View: TokenHolderTable,
  getItems: (rpcEndpoint, params, options) => getHolderForERC721Token(rpcEndpoint, params.address, params.tokenId, params.atBlock),
}

export const Actions = {
  GenerateReport,
  ContractCreationTx,
  ERC20TokenBalance,
  ERC20TokenHolders,
  ERC721HolderForToken,
  ERC721Holders,
  ERC721Tokens,
  ERC721TokensForAccount,
  Events,
  InternalToTxs,
  ToTxs,
}
