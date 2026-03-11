export const getLeaderboardQuery = /* GraphQL */ `
  query GetLeaderboard($board: String!, $window: String!, $limit: Int) {
    getLeaderboard(board: $board, window: $window, limit: $limit) {
      leaderboard {
        board
        window
        generatedAt
        entries {
          memberId
          rank
          score
          orders
          units
        }
      }
    }
  }
`;
