import numpy as np

from src.thompson import ArmResult, TrainingData, predict

actual_prob = [0.1, 0.7, 0.5]

data = TrainingData(
    Arms=[
        ArmResult(Success=0, Failures=0, ArmName="1"),
        ArmResult(Success=0, Failures=0, ArmName="2"),
        ArmResult(Success=0, Failures=0, ArmName="3"),
    ]
)

for trial in range(101):
    predicted = predict(data)

    if np.random.uniform() < actual_prob[predicted]:
        data.Arms[predicted].Success += 1
    else:
        data.Arms[predicted].Failures += 1

    # logging
    if trial % 10 == 0:
        print(f"trial: {trial}\tdata: {data}")

