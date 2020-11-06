from utils import load_table_classification_dataset

res = load_table_classification_dataset('./test.csv', ['a', 'b','c'], ['d'])
print(res)