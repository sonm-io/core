const FixedPointUtil = artifacts.require('./FixedPointUtil.sol');
const FixedPoint128 = artifacts.require('./FixedPoint128.sol');
const FixedPoint = artifacts.require('./FixedPoint.sol');
const Rating = artifacts.require('./Rating.sol');
const RatingData = artifacts.require('./RatingData.sol');

module.exports = function (deployer, network) {
    if (network === 'master' || network === 'rinkeby') {
        // rating haven't reason to deploy to mainnet or rinkeby
        return;
    }
    deployer.deploy(FixedPointUtil, { gasPrice: 0 });
    deployer.deploy(FixedPoint128, { gasPrice: 0 }).then(() => {
        deployer.deploy(RatingData, 0, { gasPrice: 0 }).then(() => {
            deployer.deploy(Rating, RatingData.address);
        });
    });
    deployer.link(FixedPoint128, Rating);
};
