# block-verifications

This project demonstrates how to verify block by block using zilliqa golang sdk. Here is the steps:

1. get current ds block number, tx block number and DS committee list
2. wait a new tx block got generated
3. check tx block's header, detect if there is a new ds block generated
4. if there is a new ds block, verify it first, get new DS committee list, verify tx block after that
5. if there is no new ds block, verify tx block directly
6. goto 2

Links

1. tx block verification: https://github.com/Zilliqa/gozilliqa-sdk/blob/master/verifier/verifier.go#L116
2. ds block verification: https://github.com/Zilliqa/gozilliqa-sdk/blob/master/verifier/verifier.go#L108
