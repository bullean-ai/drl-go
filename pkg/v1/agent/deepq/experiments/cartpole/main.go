package main

import (
	"github.com/aunum/gold/pkg/v1/agent/deepq"
	"github.com/aunum/gold/pkg/v1/common"
	"github.com/aunum/gold/pkg/v1/common/require"
	envv1 "github.com/aunum/gold/pkg/v1/env"
	"github.com/aunum/gold/pkg/v1/track"
	"github.com/aunum/log"
)

func main() {
	s, err := envv1.NewLocalServer(envv1.GymServerConfig)
	require.NoError(err)
	defer s.Close()

	env, err := s.Make("CartPole-v0",
		envv1.WithNormalizer(envv1.NewExpandDimsNormalizer(0)),
	)
	require.NoError(err)

	agent, err := deepq.NewAgent(deepq.DefaultAgentConfig, env)
	require.NoError(err)

	agent.View()

	numEpisodes := 2000
	agent.Epsilon = common.DefaultDecaySchedule(common.WithDecayRate(0.9995))
	for _, episode := range agent.MakeEpisodes(numEpisodes) {
		init, err := env.Reset()
		require.NoError(err)

		state := init.Observation

		score := episode.TrackScalar("score", 0, track.WithAggregator(track.Max))

		for _, timestep := range episode.Steps(env.MaxSteps()) {
			action, err := agent.Action(state)
			require.NoError(err)

			outcome, err := env.Step(action)
			require.NoError(err)

			score.Inc(outcome.Reward)

			event := deepq.NewEvent(state, action, outcome)
			agent.Remember(event)

			err = agent.Learn()
			require.NoError(err)

			if outcome.Done {
				log.Successf("Episode %d finished after %d timesteps", episode.I, timestep.I+1)
				break
			}
			state = outcome.Observation

			err = agent.Render(env)
			require.NoError(err)
		}
		episode.Log()
	}
	agent.Wait()
	env.End()
}
