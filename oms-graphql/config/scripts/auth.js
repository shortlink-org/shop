function requireUserId({ request }) {
  const userId = request.headers["x-user-id"];

  if (userId && userId.trim() !== "") {
    return {
      request: {
        ...request,
        headers: {
          ...request.headers,
          "user-id": userId
        }
      }
    };
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
