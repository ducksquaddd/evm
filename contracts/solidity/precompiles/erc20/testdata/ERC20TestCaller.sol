// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity >=0.8.17;

import "../IERC20Metadata.sol" as erc20Precompile;

/// @title ERC20TestCaller
/// @author Erric
/// @dev This contract is used to test external contract calls to the ERC20 precompile.
contract ERC20TestCaller {
    erc20Precompile.IERC20Metadata public token;
    uint256 public counter;

    constructor(address tokenAddress) {
        token = erc20Precompile.IERC20Metadata(tokenAddress);
        counter = 0;
    }

    function transfer(address to, uint256 amount) external returns (bool) {
        return token.transfer(to, amount);
    }

    function transferFrom(address from, address to, uint256 amount) external returns (bool) {
        return token.transferFrom(from, to, amount);
    }

    function approve(address spender, uint256 amount) external returns (bool) {
        return token.approve(spender, amount);
    }

    function allowance(address owner, address spender) external view returns (uint256) {
        return token.allowance(owner, spender);
    }

    function balanceOf(address owner) external view returns (uint256) {
        return token.balanceOf(owner);
    }

    function totalSupply() external view returns (uint256) {
        return token.totalSupply();
    }

    function name() external view returns (string memory) {
        return token.name();
    }

    function symbol() external view returns (string memory) {
        return token.symbol();
    }

    function decimals() external view returns (uint8) {
        return token.decimals();
    }

    function transferWithRevert(
        address to,
        uint256 amount,
        bool before,
        bool aft
    ) public payable returns (bool) {
        counter++;
        
        bool res = token.transfer(to, amount);

        if (before) {
            require(false, "revert here");
        }

        counter--;
        
        if (aft) {
            require(false, "revert here");
        }
        return res;
    }

    function testTransferAndSend(
        address payable _source,
        uint256 amount_to_transfer,
        uint256 amount_to_send,
        uint256 amount_to_send_after,
        bool _before,
        bool _after
    ) public payable returns (bool) {
        (bool sent, ) = _source.call{value: amount_to_send}("");
        require(sent, "Failed to send Ether to delegator");
        
        if (_before) {
            counter++;
            require(false, "revert here");
        }
        
        bool res = token.transfer(_source, amount_to_transfer);
        require(res, "Failed to send Ether to delegator");

        if (_after) {
            counter++;
            require(false, "revert here");
        }

        (sent, ) = _source.call{value: amount_to_send_after}("");
        require(sent, "Failed to send Ether to delegator");

        return sent;
    }

    function transfersWithTry(
        address payable receiver,
        uint256 amount_to_transfer,
        uint256 amount_to_fail
    ) public payable {
        counter++;
        bool res = token.transfer(receiver, amount_to_transfer);
        require(res, "fail to transfer");
        try
            ERC20TestCaller(address(this))
                .transferWithRevert(
                    receiver,                    
                    amount_to_fail,
                    true, 
                    true
                )
        {} catch {}
        counter++;
    }
}
