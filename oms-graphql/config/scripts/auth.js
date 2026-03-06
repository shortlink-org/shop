function requireUserId({ request }) {
  const userId = request.headers["user-id"];

  if (userId && userId.trim() !== "") {
    return { request };
  }

  return {
    response: {
      status: 401,
      headers: {
        "content-type": "application/json"
      },
      body: JSON.stringify({
        error: {
          message: "Missing required user-id header"
        }
      })
    }
  };
}
