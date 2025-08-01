// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.10;

import { ERC20 } from "solady/src/tokens/ERC20.sol";
import {Ownable} from "solady/src/auth/Ownable.sol";

contract MockCoin is ERC20, Ownable {

    string internal _name;
    string internal _symbol;

    constructor(string memory name_, string memory symbol_) ERC20() {
        _name = name_;
        _symbol = symbol_;
    }

    function name() public view override returns (string memory) {
        return _name;
    }

    function symbol() public view override returns (string memory) {
        return _symbol;
    }

    function mint(address account, uint256 amount) external onlyOwner {
        _mint(account, amount);
    }

    function burn(address account, uint256 amount) external {
        _burn(account, amount);
    }
}