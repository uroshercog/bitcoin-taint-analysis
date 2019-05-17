import pandas as pd

'''
Filter the JSON so we only have things we want.


Structure we have in the resulting files ("./data/bitcoin_filtered/") is the following:

[input_addr_1;ammount, input_addr_2;ammount,...]:[output_addr_1;ammount, output_addr_2;ammount,...]\n
[input_addr_1;ammount, input_addr_2;ammount,...]:[output_addr_1;ammount, output_addr_2;ammount,...]\n
...

Each row represents a transaction. Inside row there are [inputs]:[outputs].



'''
for i in range(1,9):
    print(i)
    df = pd.read_json('./data/bitcoin_1_month/chunk'+str(i)+'.json', lines=True)
    # We want 'addresses' and 'value'
    f = open("./data/bitcoin_filtered/chunk"+str(i)+".data", "w+")
    for idx, row in df.iterrows():
        # print(row[0])
        # inputs
        try:
            inputs = row[0]
            f.write("[")
            length = len(inputs)
            idx = 0
            for input in inputs:
                idx += 1
                f.write(input['addresses'][0] + ";")
                f.write(str(input['value']))
                if length != idx:
                    f.write(",")
            f.write("]")    
        except Exception as e:
            print(e)
        # outputs
        f.write(":")
        try:
            outputs = row[1]
            f.write("[")
            length = len(outputs)
            idx = 0
            for output in outputs:
                idx += 1
                f.write(output['addresses'][0] + ";")
                f.write(str(output['value']))
                if length != idx:
                    f.write(",")
            f.write("]" + "\n")   
        except Exception as e:
            print(e)