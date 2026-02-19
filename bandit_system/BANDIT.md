For each piece of music, compute the most representative Tag 

For each user, compute the weights for each Tag

Once tag is selected by BANDIT, pass to popularity system (recommend the most popular songs with tag T)


Two step
- predict (sync -> wait for service to return prediction)
- update (async -> task queue - Kafka?)